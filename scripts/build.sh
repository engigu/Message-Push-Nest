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
    local end_ms="${2:-$(get_timestamp_ms)}"
    local ms=$(( end_ms - start_ms ))

    if [ -z "${ms}" ] || [ "${ms}" -le 0 ]; then
        echo "0s"
        return
    fi

    if [ "${ms}" -lt 1000 ]; then
        echo "${ms}ms"
        return
    fi

    local sec=$(( ms / 1000 ))
    local remain_ms=$(( (ms % 1000) / 100 ))

    if [ "${sec}" -lt 60 ]; then
        if [ "${remain_ms}" -gt 0 ]; then
            echo "${sec}.${remain_ms}s"
        else
            echo "${sec}s"
        fi
        return
    fi

    local min=$(( sec / 60 ))
    local rem_sec=$(( sec % 60 ))
    echo "${min}m${rem_sec}s"
}

# ============================================================
# 基础上下文与环境变量
# ============================================================
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

VERSION="${VERSION:-}"
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
# 2. 构建单平台 Go 二进制
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

    go mod download

    # 1. 编译 linux-amd64
    build_server linux amd64 "${base_output}/bin/linux-amd64/Message-Nest"

    # 2. 编译 linux-arm64
    build_server linux arm64 "${base_output}/bin/linux-arm64/Message-Nest"

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

    log_step "开始全平台 Release 打包"
    log_info "发布包输出目录: ${pkg_dir}"
    log_info "Docker 资产目录: ${docker_dir}"
    log_info "版本号: ${VERSION}"

    mkdir -p "${pkg_dir}" "${docker_dir}/bin/linux-amd64" "${docker_dir}/bin/linux-arm64"
    echo "${VERSION}" > .release_version

    go mod download

    _pack_tar() {
        local os=$1; local arch=$2
        local bin_file="Message-Nest"
        log_info "打包 [${os}/${arch}] -> tar.gz..."
        CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" go build -ldflags="${LDFLAGS}" -o "${bin_file}" .
        tar -czf "${pkg_dir}/Message-Nest-${VERSION}-${os}-${arch}.tar.gz" "${bin_file}" conf/app.example.ini LICENSE NOTICE
        if [ "${os}" = "linux" ]; then
            cp "${bin_file}" "${docker_dir}/bin/linux-${arch}/Message-Nest"
        fi
        rm -f "${bin_file}"
    }

    # 1. Linux amd64 & arm64
    _pack_tar linux amd64
    _pack_tar linux arm64

    # 2. Windows amd64 (.zip)
    log_info "打包 [windows/amd64] -> zip..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "Message-Nest.exe" .
    zip -q -j "${pkg_dir}/Message-Nest-${VERSION}-windows-amd64.zip" "Message-Nest.exe" conf/app.example.ini LICENSE NOTICE
    rm -f "Message-Nest.exe"

    # 3. macOS Apple Silicon (arm64) & Intel (amd64)
    _pack_tar darwin arm64
    _pack_tar darwin amd64

    local duration
    duration=$(format_duration "${start_time}")
    log_step "Release 全平台构建完成 (总耗时: ${duration})，清单如下:"
    ls -lh "${pkg_dir}" | awk '{print "   " $9 " (" $5 ")"}'
}

# ============================================================
# 5. 清理临时产物
# ============================================================
clean() {
    log_step "清理构建目录与临时文件..."
    rm -rf web/dist dist-assets release-pkg Message-Nest Message-Nest.exe
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
# 命令分发入口
# ============================================================
CMD="${1:-all}"
shift || true

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
        echo "使用方法: $0 {all|web|server [os] [arch] [output]|ci-artifacts [output_dir]|release-artifacts [pkg_dir] [docker_dir]|clean}"
        exit 0
        ;;
    *)
        log_warn "未知命令: ${CMD}"
        echo "使用方法: $0 {all|web|server [os] [arch] [output]|ci-artifacts [output_dir]|release-artifacts [pkg_dir] [docker_dir]|clean}"
        exit 1
        ;;
esac
