# myy_chat Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-12-27

## Active Technologies

- **AI角色对话平台** (001-ai-companion-platform): 企业级AI角色对话系统，支持用户认证、角色管理、智能对话、记忆系统、工具能力、积分计费等功能

### Frontend Stack
- React 19.2.0 + TypeScript 5.9.3
- Vite 7.2.4 (Build Tool)
- React Router DOM 7.11.0 (Routing)
- Tailwind CSS 3.4.17 (Styling)
- Framer Motion 12.23.26 (Animation)
- Lucide React 0.562.0 (Icons)
- React Context API (State Management)
- Custom i18n (中文/英文)

### Backend Stack
- Go 1.21+ with Kratos v2
- gRPC + Protobuf (Internal Communication)
- HTTP RESTful (External API)
- Wire (Dependency Injection)

## Project Structure

```text
myy_chat/
├── frontend/                    # React SPA 前端
│   └── src/
│       ├── components/         # 可复用组件
│       │   └── layout/        # 布局组件
│       ├── context/           # React Context 状态管理
│       ├── pages/             # 页面组件
│       ├── services/          # API 服务层
│       ├── types/             # TypeScript 类型
│       └── lib/               # 工具函数
│
├── backend/golang/             # Go 微服务后端 (Kratos v2)
│   ├── api/                   # Proto 定义（服务间共享）
│   │   ├── user/v1/          # 用户服务 API
│   │   ├── character/v1/     # 角色服务 API
│   │   ├── chat/v1/          # 聊天服务 API
│   │   └── admin/v1/         # 管理后台 API
│   ├── app/                   # 微服务实现
│   │   └── {service}/
│   │       ├── cmd/          # 服务入口
│   │       └── internal/     # 内部实现
│   ├── pkg/                   # 共享库
│   ├── templates/             # 新服务模板
│   └── third_party/          # 第三方 Proto
│
└── specs/                      # 功能规格文档
    └── 001-ai-companion-platform/
        ├── spec.md            # 功能规格
        ├── plan.md            # 实现计划
        ├── tasks.md           # 任务列表
        └── data-model.md      # 数据模型
```

## Commands

### Frontend
```bash
cd frontend
npm run dev        # 启动开发服务器 (http://localhost:5173)
npm run build      # 生产构建
npm run lint       # 代码检查
npm run preview    # 预览生产构建
```

### Backend (Root Makefile)
```bash
cd backend/golang

# 初始化
make init           # 安装依赖工具

# Proto 生成
make proto          # 生成所有 proto 代码
make proto-{service} # 生成指定服务的 proto

# 构建和运行
make build          # 构建所有服务
make build-{service} # 构建指定服务
make run-{service}   # 运行指定服务

# 测试和调试
make test-{service}  # 测试指定服务
make debug-{service} # 调试模式 (Delve, port 2345)

# 新建服务
make new SERVICE=xxx # 创建新服务骨架
```

## Code Style

### Frontend
- TypeScript strict mode
- React Hooks (函数式组件)
- Tailwind CSS utility classes
- ESLint + Prettier

### Backend
- Go 官方规范
- 中文注释，英文日志
- gRPC + Protobuf

## Recent Changes

- 001-ai-companion-platform: 添加完整技术选型
- Frontend: React 19 + TypeScript + Vite 骨架实现
- Backend: Kratos v2 微服务脚手架

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
