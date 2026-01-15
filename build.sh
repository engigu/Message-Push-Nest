#!/bin/bash

# Message Nest 构建脚本
# 用于手动构建前后端应用

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
BUILD_MODE="prod"  # 构建模式: prod 或 demo
USE_NPM_CI=false   # 是否使用 npm ci

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印分隔线
print_separator() {
    echo "=================================================="
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未安装，请先安装 $1"
        exit 1
    fi
}

# 清理旧的构建文件
clean_build() {
    print_info "清理旧的构建文件..."
    
    # 清理前端构建文件
    if [ -d "web/dist" ]; then
        rm -rf web/dist
        print_success "已清理前端构建文件"
    fi
    
    # 清理后端构建文件
    if [ -f "message-nest" ]; then
        rm -f message-nest
        print_success "已清理后端构建文件"
    fi
    
    if [ -f "Message-Nest" ]; then
        rm -f Message-Nest
        print_success "已清理后端构建文件"
    fi
    
    if [ -f "message-nest.exe" ]; then
        rm -f message-nest.exe
        print_success "已清理 Windows 构建文件"
    fi
    
    # 清理多平台构建目录
    if [ -d "build" ]; then
        rm -rf build
        print_success "已清理多平台构建目录"
    fi
}

# 构建前端
build_frontend() {
    print_separator
    print_info "开始构建前端..."
    
    # 检查 Node.js 和 npm
    check_command "node"
    check_command "npm"
    
    print_info "Node 版本: $(node --version)"
    print_info "npm 版本: $(npm --version)"
    
    # 进入前端目录
    cd web
    
    # 安装依赖
    if [ "$USE_NPM_CI" = true ]; then
        print_info "使用 npm ci 安装依赖（CI 模式）..."
        npm ci
    else
        if [ ! -d "node_modules" ]; then
            print_info "安装前端依赖..."
            npm install
        else
            print_info "前端依赖已存在，跳过安装"
        fi
    fi
    
    # 构建前端
    if [ "$BUILD_MODE" = "demo" ]; then
        print_info "打包前端资源（Demo 模式）..."
        npm run build:demo
    else
        print_info "打包前端资源（生产模式）..."
        export NODE_ENV=prod
        npm run build
    fi
    
    # 返回根目录
    cd ..
    
    # 检查构建结果
    if [ -d "web/dist" ]; then
        print_success "前端构建完成！"
        print_info "构建文件位置: web/dist/"
        
        # 显示构建文件大小
        DIST_SIZE=$(du -sh web/dist | cut -f1)
        print_info "构建文件大小: ${DIST_SIZE}"
    else
        print_error "前端构建失败！"
        exit 1
    fi
}

# 构建后端
build_backend() {
    print_separator
    print_info "开始构建后端..."
    
    # 检查 Go
    check_command "go"
    
    print_info "Go 版本: $(go version)"
    
    # 下载依赖
    print_info "整理 Go 依赖..."
    go mod tidy
    
    # 构建后端（嵌入前端资源）
    print_info "编译后端程序（嵌入前端资源）..."
    
    # 获取构建信息
    BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
    
    print_info "版本: ${VERSION}"
    print_info "提交: ${GIT_COMMIT}"
    print_info "构建时间: ${BUILD_TIME}"
    
    # 构建参数
    LDFLAGS="-s -w"
    LDFLAGS="${LDFLAGS} -X 'main.Version=${VERSION}'"
    LDFLAGS="${LDFLAGS} -X 'main.BuildTime=${BUILD_TIME}'"
    LDFLAGS="${LDFLAGS} -X 'main.GitCommit=${GIT_COMMIT}'"
    
    # 根据操作系统构建
    OS=$(uname -s)
    case "$OS" in
        Linux*)
            print_info "检测到 Linux 系统，构建 Linux 版本..."
            CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o Message-Nest main.go
            ;;
        Darwin*)
            print_info "检测到 macOS 系统，构建 macOS 版本..."
            go build -ldflags "${LDFLAGS}" -o Message-Nest main.go
            ;;
        MINGW*|MSYS*|CYGWIN*)
            print_info "检测到 Windows 系统，构建 Windows 版本..."
            go build -ldflags "${LDFLAGS}" -o message-nest.exe main.go
            ;;
        *)
            print_warning "未知操作系统: $OS，使用默认构建..."
            go build -ldflags "${LDFLAGS}" -o Message-Nest main.go
            ;;
    esac
    
    # 检查构建结果
    if [ -f "Message-Nest" ] || [ -f "message-nest.exe" ]; then
        print_success "后端构建完成！"
        
        # 显示文件信息
        if [ -f "Message-Nest" ]; then
            FILE_SIZE=$(du -h Message-Nest | cut -f1)
            print_info "可执行文件: Message-Nest (${FILE_SIZE})"
        fi
        
        if [ -f "message-nest.exe" ]; then
            FILE_SIZE=$(du -h message-nest.exe | cut -f1)
            print_info "可执行文件: message-nest.exe (${FILE_SIZE})"
        fi
    else
        print_error "后端构建失败！"
        exit 1
    fi
}

# 构建多平台版本
build_cross_platform() {
    print_separator
    print_info "开始构建多平台版本..."
    
    # 创建构建目录
    BUILD_DIR="build"
    mkdir -p ${BUILD_DIR}
    
    # 构建信息
    BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S')
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
    
    LDFLAGS="-s -w"
    LDFLAGS="${LDFLAGS} -X 'main.Version=${VERSION}'"
    LDFLAGS="${LDFLAGS} -X 'main.BuildTime=${BUILD_TIME}'"
    LDFLAGS="${LDFLAGS} -X 'main.GitCommit=${GIT_COMMIT}'"
    
    # 构建 Linux amd64
    print_info "构建 Linux amd64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/message-nest-linux-amd64 main.go
    
    # 构建 Linux arm64
    print_info "构建 Linux arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/message-nest-linux-arm64 main.go
    
    # 构建 macOS amd64
    print_info "构建 macOS amd64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/message-nest-darwin-amd64 main.go
    
    # 构建 macOS arm64 (Apple Silicon)
    print_info "构建 macOS arm64..."
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/message-nest-darwin-arm64 main.go
    
    # 构建 Windows amd64
    print_info "构建 Windows amd64..."
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/message-nest-windows-amd64.exe main.go
    
    print_success "多平台构建完成！"
    print_info "构建文件位置: ${BUILD_DIR}/"
    ls -lh ${BUILD_DIR}/
}

# 显示帮助信息
show_help() {
    echo "Message Nest 构建脚本"
    echo ""
    echo "用法: ./build.sh [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -c, --clean             清理构建文件"
    echo "  -f, --frontend          仅构建前端"
    echo "  -b, --backend           仅构建后端"
    echo "  -a, --all               构建前后端（默认）"
    echo "  -x, --cross-platform    构建多平台版本"
    echo "  -d, --demo              使用 Demo 模式构建前端"
    echo "  --ci                    使用 CI 模式（npm ci）"
    echo ""
    echo "示例:"
    echo "  ./build.sh              # 构建前后端（生产模式）"
    echo "  ./build.sh -d           # 构建前后端（Demo 模式）"
    echo "  ./build.sh -f           # 仅构建前端"
    echo "  ./build.sh -b           # 仅构建后端"
    echo "  ./build.sh -c           # 清理构建文件"
    echo "  ./build.sh -x           # 构建多平台版本"
    echo "  ./build.sh --ci         # CI 模式构建"
}

# 主函数
main() {
    print_separator
    print_info "Message Nest 构建脚本"
    print_separator
    
    # 解析参数
    BUILD_FRONTEND=false
    BUILD_BACKEND=false
    BUILD_CROSS=false
    
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--clean)
                clean_build
                print_success "清理完成！"
                exit 0
                ;;
            -f|--frontend)
                BUILD_FRONTEND=true
                shift
                ;;
            -b|--backend)
                BUILD_BACKEND=true
                shift
                ;;
            -x|--cross-platform)
                BUILD_CROSS=true
                shift
                ;;
            -d|--demo)
                BUILD_MODE="demo"
                print_info "使用 Demo 模式"
                shift
                ;;
            --ci)
                USE_NPM_CI=true
                print_info "使用 CI 模式"
                shift
                ;;
            -a|--all)
                BUILD_FRONTEND=true
                BUILD_BACKEND=true
                shift
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定任何构建选项，默认构建全部
    if [ "$BUILD_FRONTEND" = false ] && [ "$BUILD_BACKEND" = false ] && [ "$BUILD_CROSS" = false ]; then
        BUILD_FRONTEND=true
        BUILD_BACKEND=true
    fi
    
    # 执行构建
    if [ "$BUILD_FRONTEND" = true ]; then
        build_frontend
    fi
    
    if [ "$BUILD_CROSS" = true ]; then
        build_cross_platform
    elif [ "$BUILD_BACKEND" = true ]; then
        build_backend
    fi
    
    print_separator
    print_success "构建完成！"
    print_separator
    
    # 显示运行提示
    if [ -f "Message-Nest" ]; then
        echo ""
        print_info "运行应用:"
        echo "  ./Message-Nest"
        echo ""
    fi
    
    if [ -f "message-nest.exe" ]; then
        echo ""
        print_info "运行应用:"
        echo "  message-nest.exe"
        echo ""
    fi
}

# 执行主函数
main "$@"
