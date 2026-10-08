#!/usr/bin/env bash
set -eo pipefail

# ============================================================
# 日志与格式化输出
# ============================================================
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[1;36m"
C_GREEN="\033[1;32m"
C_YELLOW="\033[1;33m"
C_BLUE="\033[1;34m"
C_RED="\033[1;31m"

log_info() {
    printf "${C_CYAN}[Build]${C_RESET} %s\n" "$*"
}

log_step() {
    printf "\n${C_BLUE}==>${C_RESET} ${C_BOLD}%s${C_RESET}\n" "$*"
}

log_success() {
    printf "${C_GREEN}✓ %s${C_RESET}\n" "$*"
}

log_warn() {
    printf "${C_YELLOW}⚠ %s${C_RESET}\n" "$*"
}

log_error() {
    printf "${C_RED}✗ %s${C_RESET}\n" "$*"
}

get_file_size() {
    ls -lh "$1" 2>/dev/null | awk '{print $5}' || echo ""
}

get_timestamp_ms() {
    local ns
    ns=$(date +%s%N 2>/dev/null || true)
    if [[ "${ns}" =~ ^[0-9]{19}$ ]]; then
        echo "${ns:0:13}"
    else
        echo "$(( $(date +%s) * 1000 ))"
    fi
}

format_duration() {
    local start_ms=$1
    local end_ms
    end_ms=$(get_timestamp_ms)
    local diff_ms=$((end_ms - start_ms))
    if [ ${diff_ms} -lt 1000 ]; then
        echo "${diff_ms}ms"
    else
        local sec=$((diff_ms / 1000))
        local ms=$((diff_ms % 1000))
        echo "${sec}.${ms}s"
    fi
}

# ============================================================
# 项目基础路径与版本管理
# ============================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

if [ -z "${VERSION}" ]; then
    if [ -f .release_version ]; then
        VERSION="$(cat .release_version | tr -d '\r\n')"
    elif git describe --tags --always 2>/dev/null; then
        VERSION="$(git describe --tags --always)"
    else
        VERSION="default"
    fi
fi

BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
LDFLAGS="-s -w"

# ============================================================
# 1. 构建 Web 前端
# ============================================================
build_web() {
    local start_time
    start_time=$(get_timestamp_ms)
    log_step "构建前端静态资源 (Web)"

    if [ ! -d "web" ]; then
        log_error "未找到 web 目录"
        exit 1
    fi

    cd web
    log_info "安装前端依赖 (npm ci)..."
    npm ci
    log_info "执行前端打包 (npm run build)..."
    export NODE_ENV=prod
    npm run build
    cd "${ROOT_DIR}"

    local duration
    duration=$(format_duration "${start_time}")
    log_success "前端资源构建完成 (耗时: ${duration})"
}

# ============================================================
# 2. 编译 Go 服务端 (通用单平台编译)
# ============================================================
build_server() {
    local os="${1:-linux}"
    local arch="${2:-amd64}"
    local output="${3:-Message-Nest}"
    local start_time
    start_time=$(get_timestamp_ms)

    log_step "编译 Go 服务端 [${os}/${arch}] -> ${output}"

    mkdir -p "$(dirname "${output}")"
    echo "${VERSION}" > .release_version

    CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
        go build -ldflags="${LDFLAGS}" -o "${output}" .

    local duration
    duration=$(format_duration "${start_time}")
    local fsize
    fsize=$(get_file_size "${output}")
    log_success "编译完成 -> ${output} (大小: ${fsize:-未知}, 耗时: ${duration})"
}

# ============================================================
# 3. CI 部署打包产物（专供 .github/workflows/deploy.yml）
# ============================================================
build_ci_artifacts() {
    local base_output="${1:-dist-assets}"
    local start_time
    start_time=$(get_timestamp_ms)

    log_step "开始打包 CI 部署产物 (双架构)"
    log_info "目标目录: ${base_output}"
    log_info "版本标记: ${VERSION}"

    mkdir -p "${base_output}/bin/linux-amd64" "${base_output}/bin/linux-arm64"
    echo "${VERSION}" > .release_version

    log_info "预下载模块依赖 (go mod download)..."
    go mod download

    # 1. 编译 linux-amd64
    build_server linux amd64 "${base_output}/bin/linux-amd64/Message-Nest"

    # 2. 编译 linux-arm64
    build_server linux arm64 "${base_output}/bin/linux-arm64/Message-Nest"

    log_step "CI 部署产物打包完成，清单如下:"
    find "${base_output}" -type f -exec ls -lh {} + | awk '{print "   " $9 " (" $5 ")"}'

    local duration
    duration=$(format_duration "${start_time}")
    log_success "CI 部署产物全部准备完成 (总耗时: ${duration})"
}

# ============================================================
# 4. Release 全平台发布打包（专供 .github/workflows/release.yml）
# ============================================================
build_release_artifacts() {
    local pkg_dir="${1:-release-pkg}"
    local docker_dir="${2:-dist-assets}"
    local start_time
    start_time=$(get_timestamp_ms)

    log_step "开始全平台 Release 打包 (对齐 GoReleaser 完整矩阵)"
    log_info "发布包输出目录: ${pkg_dir}"
    log_info "Docker 资产目录: ${docker_dir}"
    log_info "版本号: ${VERSION}"

    mkdir -p "${pkg_dir}" "${docker_dir}/bin/linux-amd64" "${docker_dir}/bin/linux-arm64"
    echo "${VERSION}" > .release_version

    log_info "预下载模块依赖 (go mod download)..."
    go mod download

    _pack_archive() {
        local os=$1
        local arch=$2
        local goarm=$3
        local title_os=$4
        local arch_suffix=$5
        local is_zip=$6

        local bin_name="Message-Nest"
        if [ "${os}" = "windows" ]; then
            bin_name="Message-Nest.exe"
        fi

        local archive_name="Message-Nest_${title_os}_${arch_suffix}"
        local start_pack_time
        start_pack_time=$(get_timestamp_ms)

        log_info "编译并打包 [${os}/${arch}${goarm:+v${goarm}}] -> ${archive_name}..."

        local env_args=(CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}")
        if [ -n "${goarm}" ]; then
            env_args+=(GOARM="${goarm}")
        fi

        env "${env_args[@]}" go build -ldflags="${LDFLAGS}" -o "${bin_name}" .

        if [ "${is_zip}" = "true" ]; then
            zip -q -r "${pkg_dir}/${archive_name}.zip" "${bin_name}" LICENSE README.md conf
        else
            tar -czf "${pkg_dir}/${archive_name}.tar.gz" "${bin_name}" LICENSE README.md conf
        fi

        # 同步一份供 docker 构建
        if [ "${os}" = "linux" ] && [ "${arch}" = "amd64" ]; then
            cp "${bin_name}" "${docker_dir}/bin/linux-amd64/Message-Nest"
        elif [ "${os}" = "linux" ] && [ "${arch}" = "arm64" ]; then
            cp "${bin_name}" "${docker_dir}/bin/linux-arm64/Message-Nest"
        fi

        rm -f "${bin_name}"

        local pack_duration
        pack_duration=$(format_duration "${start_pack_time}")
        local fsize
        if [ "${is_zip}" = "true" ]; then
            fsize=$(get_file_size "${pkg_dir}/${archive_name}.zip")
        else
            fsize=$(get_file_size "${pkg_dir}/${archive_name}.tar.gz")
        fi
        log_success "打包完成 -> ${archive_name} (大小: ${fsize:-未知}, 耗时: ${pack_duration})"
    }

    # 1. Darwin (macOS Apple Silicon & Intel)
    _pack_archive darwin arm64 "" Darwin arm64 false
    _pack_archive darwin amd64 "" Darwin x86_64 false

    # 2. FreeBSD
    _pack_archive freebsd arm64 "" Freebsd arm64 false
    _pack_archive freebsd amd64 "" Freebsd x86_64 false

    # 3. Linux (amd64, arm64, armv7)
    _pack_archive linux arm64 "" Linux arm64 false
    _pack_archive linux arm 7 Linux armv7 false
    _pack_archive linux amd64 "" Linux x86_64 false

    # 4. OpenBSD
    _pack_archive openbsd arm64 "" Openbsd arm64 false
    _pack_archive openbsd amd64 "" Openbsd x86_64 false

    # 5. Windows (arm64, x86_64)
    _pack_archive windows arm64 "" Windows arm64 true
    _pack_archive windows amd64 "" Windows x86_64 true

    # 6. 生成 SHA256 校验和 (checksums.txt)
    log_step "生成 SHA256 校验和文件 (checksums.txt)..."
    (
        cd "${pkg_dir}"
        if command -v sha256sum >/dev/null 2>&1; then
            sha256sum Message-Nest_* > checksums.txt
        elif command -v shasum >/dev/null 2>&1; then
            shasum -a 256 Message-Nest_* > checksums.txt
        fi
    )
    log_success "校验和生成完成 -> checksums.txt"

    local total_duration
    total_duration=$(format_duration "${start_time}")
    log_step "Release 全平台构建完成 (总耗时: ${total_duration})，清单如下:"
    ls -lh "${pkg_dir}" | awk '{print "   " $9 " (" $5 ")"}'
}

# ============================================================
# 5. 清理临时产物
# ============================================================
clean() {
    log_step "清理构建目录与临时文件..."
    rm -rf web/dist dist-assets release-pkg Message-Nest Message-Nest.exe .release_version
    log_success "清理完成"
}

# ============================================================
# 6. 一键全量自构建 (本地开发)
# ============================================================
build_all() {
    clean
    build_web
    build_server
    log_success "全量构建完成，可直接执行 ./Message-Nest 启动应用！"
}

# ============================================================
# 命令调度入口
# ============================================================
CMD="${1:-server}"
shift || true

echo "============================================================"
echo "           消息推送中心 (Message-Push-Nest) 构建系统"
echo "============================================================"
log_info "当前指令 : ${CMD}"
log_info "项目根目录: ${ROOT_DIR}"
log_info "版本标记 : ${VERSION}"
log_info "构建时间 : ${BUILD_TIME}"
log_info "Go 环境  : $(go version 2>/dev/null || echo '未检测到 Go 环境')"
echo "============================================================"

case "${CMD}" in
    web)
        build_web "$@"
        ;;
    server)
        build_server "$@"
        ;;
    ci-artifacts)
        build_ci_artifacts "$@"
        ;;
    release-artifacts)
        build_release_artifacts "$@"
        ;;
    clean)
        clean "$@"
        ;;
    all)
        build_all "$@"
        ;;
    help|--help|-h)
        echo "使用方法: $0 {web|server [os] [arch] [output]|ci-artifacts [output_dir]|release-artifacts [pkg_dir] [docker_dir]|clean|all}"
        exit 0
        ;;
    *)
        log_warn "未知指令: ${CMD}"
        echo "使用方法: $0 {web|server [os] [arch] [output]|ci-artifacts [output_dir]|release-artifacts [pkg_dir] [docker_dir]|clean|all}"
        exit 1
        ;;
esac
