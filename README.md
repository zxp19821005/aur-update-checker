# AUR更新检查器

一个用于检查AUR软件包上游版本更新的工具，可以帮助您更有效地管理AUR软件包。

## 功能特点

- 添加、编辑、删除软件包
- 自动从AUR获取当前版本和更新日期
- 根据上游URL和版本提取关键字获取上游最新版本
- 支持定时检查功能
- 软件界面支持列表显示和卡片显示
- 分级日志系统，支持彩色显示
- 数据库存储软件包信息、AUR信息和上游信息
- 支持检查测试版本功能
- 多种上游检查器支持（GitHub、GitLab、Gitee、HTTP、JSON、NPM、PyPI等）

## 技术栈

- 后端：Go + Gin + GORM + SQLite
- 前端：Vue 3 + Ant Design Vue + Pinia + Axios

## 数据库设计

### 基本信息表 (packageInfo)

| 字段名 | 数据类型 | 约束 | 描述 |
|--------|----------|------|------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT | 自增长主键索引 |
| name | TEXT | NOT NULL UNIQUE | 软件名称 |
| upstreamUrl | TEXT | NOT NULL | 上游URL |
| versionExtractKey | TEXT | NOT NULL | 版本提取关键字 |
| checkTestVersion | INTEGER | NOT NULL DEFAULT 0 | 是否检查测试版本(0:不检查,1:检查) |

### AUR信息表 (aurInfo)

| 字段名 | 数据类型 | 约束 | 描述 |
|--------|----------|------|------|
| packageId | INTEGER | PRIMARY KEY, FOREIGN KEY | 关联packageInfo的id |
| aurVersion | TEXT | NOT NULL | AUR版本 |
| upstreamVersionRef | TEXT | NOT NULL | 上游版本提取参考值 |
| aurCreateDate | DATETIME | NOT NULL | AUR创建日期 |
| aurUpdateDate | DATETIME | NOT NULL | AUR更新日期 |
| aurUpdateState | INTEGER | NOT NULL DEFAULT 0 | AUR更新状态(0:未检查,1:成功,2:失败) |

### 上游信息表 (upstreamInfo)

| 字段名 | 数据类型 | 约束 | 描述 |
|--------|----------|------|------|
| packageId | INTEGER | PRIMARY KEY, FOREIGN KEY | 关联packageInfo的id |
| upstreamVersion | TEXT | NOT NULL | 上游版本 |
| upstreamUpdateDate | DATETIME | NOT NULL | 上游版本更新日期 |
| upstreamUpdateState | INTEGER | NOT NULL DEFAULT 0 | 上游版本更新状态(0:未检查,1:成功,2:失败) |

## 安装与运行

### 环境要求

- Go 1.21+
- Node.js 16+

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/zxp19821005/aur-update-checker.git
cd aur-update-checker
```

2. 后端
   - 安装依赖：
     ```bash
     go mod download
     ```
   - 构建并运行：
     ```bash
     go run cmd/main.go
     ```
   后端服务将在 `http://localhost:8080` 上运行。

3. 前端
   - 进入前端目录：
     ```bash
     cd frontend
     ```
   - 安装依赖：
     ```bash
     npm install
     ```
   - 运行开发服务器：
     ```bash
     npm run dev
     ```
   前端应用将在 `http://localhost:5173` 上运行。

## 使用说明

### 添加软件包

1. 点击"软件包管理"页面中的"添加软件包"按钮
2. 填写软件包名称、上游URL和版本提取关键字
3. 点击"确定"保存

### 检查版本更新

1. 在"软件包管理"页面中，可以单独检查某个软件包的AUR版本或上游版本
2. 也可以批量检查所有软件包的AUR版本或上游版本
3. 检查结果会显示在软件包列表中

### 设置定时任务

1. 在"设置"页面中，可以设置定时检查任务
2. 设置检查间隔（分钟）
3. 点击"启动定时任务"按钮

### 查看日志

1. 在"日志"页面中，可以查看应用运行日志
2. 可以按日志级别筛选
3. 支持自动刷新功能

## 开发说明

### 目录结构

```
aur-update-checker/
├── cmd/                   # 命令行目录
│   └── main.go           # 程序入口
├── internal/             # 内部包
│   ├── checkers/         # 上游检查器
│   ├── config/           # 配置文件
│   ├── container/        # 依赖注入容器
│   ├── database/         # 数据库相关
│   ├── errors/           # 错误处理
│   ├── handlers/         # 请求处理器
│   ├── interfaces/       # 接口定义
│   ├── logger/           # 日志系统
│   ├── server/           # HTTP服务器
│   ├── services/         # 业务逻辑
│   └── utils/           # 工具函数
├── frontend/             # 前端代码
│   ├── dist/             # 构建后的前端文件
│   ├── public/           # 静态资源
│   ├── src/              # 前端源代码
│   │   ├── assets/       # 静态资源
│   │   ├── components/   # 组件
│   │   ├── models/       # 前端数据模型
│   │   ├── services/     # 前端服务
│   │   ├── stores/       # 状态管理
│   │   ├── utils/        # 工具函数
│   │   ├── views/        # 页面视图
│   │   ├── App.vue       # 主组件
│   │   └── main.js       # 前端入口
│   ├── index.html        # HTML模板
│   ├── package.json      # 前端依赖
│   └── vite.config.js    # Vite配置
├── go.mod                # Go模块文件
├── go.sum                # Go模块校验和
└── README.md             # 项目说明文档
```

### 核心模块

#### 数据库模块

- `internal/database/db_models.go`: 定义数据模型
- `internal/database/db_handler.go`: 数据库连接与操作

#### 日志模块

- `internal/logger/app_logger.go`: 实现分级日志系统，支持彩色输出
- `internal/logger/context_logger.go`: 上下文日志记录器

#### 服务层

- `internal/services/pkg_info_service.go`: 软件包信息服务
- `internal/services/aur_service.go`: AUR服务
- `internal/services/upstream_version_checker_service.go`: 上游版本检查服务
- `internal/services/update_timer_service.go`: 定时任务服务
- `internal/services/log_service.go`: 日志服务

#### 处理器层

- `internal/server/package_handler.go`: 软件包处理器
- `internal/server/api_aur_handler.go`: AUR处理器
- `internal/server/upstream_handler.go`: 上游处理器
- `internal/server/timer_handler.go`: 定时任务处理器

#### 上游检查器

- `internal/checkers/upstream_checker_registry.go`: 上游检查器注册表
- `internal/checkers/upstream_http_checker.go`: HTTP检查器
- `internal/checkers/upstream_json_checker.go`: JSON检查器
- `internal/checkers/upstream_github_checker.go`: GitHub检查器
- `internal/checkers/upstream_gitlab_checker.go`: GitLab检查器
- `internal/checkers/upstream_gitee_checker.go`: Gitee检查器
- `internal/checkers/upstream_npm_checker.go`: NPM检查器
- `internal/checkers/upstream_pypi_checker.go`: PyPI检查器
- `internal/checkers/upstream_redirect_checker.go`: 重定向检查器
- `internal/checkers/upstream_playwright_checker.go`: Playwright检查器

#### 工具函数

- `internal/utils/request_utils.go`: HTTP请求工具
- `internal/utils/version_compare.go`: 版本比较工具
- `internal/utils/config_loader.go`: 配置加载工具
- `internal/utils/error_utils.go`: 错误处理工具

## 贡献指南

欢迎提交Issue和Pull Request！

## 许可证

MIT License
