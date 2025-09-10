#!/bin/bash

# AUR更新检查器 Web 版本构建脚本
# 此脚本用于构建 Web 版本的 AUR 更新检查器

# 设置颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 打印带颜色的信息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go未安装，请先安装Go"
        exit 1
    fi
    print_info "Go版本: $(go version)"
}

# 检查pnpm是否安装
check_pnpm() {
    if ! command -v pnpm &> /dev/null; then
        print_error "pnpm未安装，请先安装pnpm"
        exit 1
    fi
    print_info "pnpm版本: $(pnpm --version)"
}

# 设置环境变量
setup_env() {
    print_info "设置环境变量..."

    # 设置Go模块代理，使用国内镜像加速依赖下载
    export GOPROXY=https://goproxy.cn,direct
    print_info "Go模块代理: $GOPROXY"
}

# 清理构建缓存
clean_cache() {
    print_info "清理构建缓存..."

    # 清理Go缓存
    go clean -cache -modcache -testcache

    # 清理前端缓存
    cd frontend
    pnpm store prune
    rm -rf node_modules .vite dist
    cd ..

    # 清理构建目录
    rm -rf dist

    print_info "构建缓存已清理"
}

# 下载依赖
download_deps() {
    print_info "下载项目依赖..."

    # 下载Go依赖
    go mod tidy

    # 下载前端依赖
    cd frontend
    pnpm install
    cd ..

    print_info "依赖下载完成"
}

# 构建前端
build_frontend() {
    print_info "构建前端..."

    cd frontend
    pnpm build
    cd ..

    # 创建静态文件目录
    mkdir -p dist/static

    # 复制前端构建文件
    cp -r frontend/dist/* dist/static/

    print_info "前端构建完成"
}

# 构建后端
build_backend() {
    print_info "构建后端..."

    # 创建dist目录
    mkdir -p dist

    # 构建Go程序
    go build -ldflags="-s -w" -o dist/aur-update-checker-web .

    # 创建数据目录
    mkdir -p dist/data

    # 创建配置文件
    if [ ! -f "dist/config.json" ]; then
        cp config.example.json dist/config.json
        print_info "已创建默认配置文件: dist/config.json"
    fi

    print_info "后端构建完成"
}

# 开发模式
run_dev() {
    print_info "启动开发模式..."

    # 启动前端开发服务器
    cd frontend
    pnpm dev &
    FRONTEND_PID=$!
    cd ..

    # 启动后端开发服务器
    go run main.go &
    BACKEND_PID=$!

    print_info "前端开发服务器已启动: http://localhost:5173"
    print_info "后端开发服务器已启动: http://localhost:8080"
    print_info "按 Ctrl+C 停止服务器"

    # 等待中断信号
    trap "kill $FRONTEND_PID $BACKEND_PID; exit" INT
    wait
}

# 生产构建
build_prod() {
    print_info "开始生产构建..."

    # 构建前端
    build_frontend

    # 构建后端
    build_backend

    # 创建启动脚本
    cat > dist/start.sh << 'EOF'
#!/bin/bash
# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 启动程序
cd "$SCRIPT_DIR"
./aur-update-checker-web
EOF
    chmod +x dist/start.sh

    print_info "生产构建完成！"
    print_info "可执行文件: dist/aur-update-checker-web"
    print_info "启动脚本: dist/start.sh"
}

# 显示帮助信息
show_help() {
    echo "AUR更新检查器 Web 版本构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  dev     启动开发模式"
    echo "  build   生产构建"
    echo "  frontend  只构建前端"
    echo "  backend   只构建后端"
    echo "  clean   清理构建缓存"
    echo "  deps    只下载依赖"
    echo "  help    显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 dev     # 启动开发模式"
    echo "  $0 build   # 生产构建"
    echo "  $0 clean   # 清理缓存"
}

# 主函数
main() {
    local action=${1:-help}

    print_info "AUR更新检查器 Web 版本构建脚本启动..."

    case $action in
        "dev")
            check_go
            check_pnpm
            setup_env
            download_deps
            run_dev
            ;;
        "build")
            check_go
            check_pnpm
            setup_env
            download_deps
            build_prod
            ;;
        "frontend")
            check_pnpm
            build_frontend
            ;;
        "backend")
            check_go
            setup_env
            build_backend
            ;;
        "clean")
            clean_cache
            ;;
        "deps")
            check_go
            check_pnpm
            setup_env
            download_deps
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知选项: $action"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
