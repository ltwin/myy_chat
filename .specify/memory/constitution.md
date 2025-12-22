<!-- Sync Impact Report:
Version change: 1.1.0 → 1.2.0 (MINOR - new development standard added)
Modified principles: N/A (existing principles unchanged)
Added sections:
  - 代码检视与提交规范 under 开发规范
Removed sections: N/A
Templates requiring updates:
  - ✅ plan-template.md - Constitution Check section needs review checkpoint gate
  - ✅ spec-template.md - Quality Requirements section already references constitution
  - ✅ tasks-template.md - Quality Tasks section already references constitution
Follow-up TODOs: N/A
-->

# MyY Chat Constitution

## Core Principles

### I. 代码质量原则
每个微服务必须有单元测试和集成测试，测试覆盖率至少达到70%。必须遵循防御式编程范式，在开发早期识别并处理潜在的安全风险、并发安全和幂等性问题。所有代码提交前必须通过静态代码分析。

### II. API规范原则
服务间通信优先使用gRPC + Protobuf（遵循Kratos框架默认规范）。对外提供的HTTP API必须严格遵循RESTful设计规范。所有API必须有完整的文档说明，包括请求/响应格式、错误码说明和使用示例。

### III. 部署原则
支持两种部署模式：Docker Compose模式用于开发环境和单机生产环境，Kubernetes模式用于云原生生产环境。配置必须支持在不同模式间灵活切换，CI/CD流水线必须同时支持两种部署方式的自动化构建和部署。

### IV. 设计原则
严格遵循面向对象设计原则，包括里氏替换法则、依赖倒置原则等。可以适当使用二十三种设计模式，但禁止过度设计。每个组件都应该有明确的单一职责，接口设计要保持简洁和一致性。

### V. 项目结构原则
所有后端服务代码统一放置在backend目录中，其中golang代码和python代码必须分别放在独立的子目录中。前端代码统一放置在frontend目录中。每个服务目录都应该包含独立的配置文件、测试文件和部署文件。

### VI. 架构原则
系统设计必须具备优秀的可读性、可扩展性和可维护性。需要考虑业务扩张后的横向扩展能力和分布式部署需求，同时也要支持简单的单机部署场景。服务间耦合度必须保持在最低水平。

## 开发规范

### 编码语言规范
- 使用中文编写代码注释，确保团队成员能够理解
- 使用英文编写日志信息，便于国际化运维和问题排查
- 在项目沟通中使用中文进行交流

### 文档规范
- 所有公共接口必须有完整的API文档
- 关键业务逻辑必须有详细的设计文档
- 部署和运维必须有完整的操作手册

### Git分支管理规范
项目必须遵循GitFlow工作流规范进行版本控制和分支管理。

**主要分支**:
- `main`: 生产环境分支，存放稳定可发布的代码，仅接受来自release或hotfix分支的合并
- `develop`: 开发主分支，包含最新的开发功能，是feature分支的基础

**支持分支**:
- `feature/*`: 功能开发分支，从develop分支创建，完成后合并回develop
  - 命名规范: `feature/功能描述` 或 `feature/issue-编号-功能描述`
- `release/*`: 发布准备分支，从develop创建，完成后同时合并到main和develop
  - 命名规范: `release/vX.Y.Z`
- `hotfix/*`: 紧急修复分支，从main创建，完成后同时合并到main和develop
  - 命名规范: `hotfix/vX.Y.Z` 或 `hotfix/问题描述`

**合并规则**:
- feature → develop: 使用Pull Request进行代码审查后合并
- release → main, develop: 发布完成后标记版本Tag
- hotfix → main, develop: 紧急修复后立即合并并标记版本Tag

**提交规范**:
- 提交信息必须清晰描述改动内容
- 建议使用Conventional Commits规范（feat:, fix:, docs:, refactor:, test:等）

### 技术文档查阅规范
在使用第三方库、框架或语言特性时，必须通过Context7 MCP工具查阅最新文档。

**强制要求**:
- 引入新的第三方库之前，必须先通过`context7.resolve-library-id`获取库ID
- 使用`context7.get-library-docs`获取最新的官方文档
- 使用不熟悉的语言特性或API时，必须先查阅对应文档

**适用场景**:
- 选择或评估第三方库时
- 配置框架（如Kratos、FastAPI等）时
- 使用数据库驱动或ORM时
- 集成外部服务或SDK时
- 使用语言高级特性时

**操作流程**:
1. 确定需要查阅的库或技术名称
2. 调用`resolve-library-id`获取Context7兼容的库ID
3. 调用`get-library-docs`获取相关主题的文档
4. 基于官方文档进行开发决策和实现

**目的**: 确保代码实现符合最新最佳实践，避免使用已废弃的API或错误的用法。

### 代码检视与提交规范
在完成一个相对完整的功能迭代或修复一个完整的Bug后，必须暂停开发流程，进行代码检视并生成规范的提交信息。

**强制暂停点**:
- 完成一个小迭代（小功能点的完整实现）
- 完成一个相对完整的功能模块
- 完成一个Bug的完整修复
- 完成一组相关的代码重构

**暂停时必须执行的操作**:
1. 停止继续开发，等待用户进行代码检视
2. 提供本次修改的简要说明
3. 列出本次修改涉及的文件清单
4. 生成符合Conventional Commits规范的git commit信息

**Commit信息格式**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**type类型**:
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档变更
- `style`: 代码格式（不影响代码运行的变动）
- `refactor`: 重构（既不是新增功能，也不是修复Bug）
- `perf`: 性能优化
- `test`: 增加测试
- `chore`: 构建过程或辅助工具的变动

**目的**: 确保代码质量通过人工检视，保持提交历史清晰可追溯，便于后续代码审查和问题定位。

## Governance

本宪法在项目开发过程中具有最高约束力，所有其他规范和流程都必须遵循宪法原则。

**版本控制策略**：
- 主版本号：向后不兼容的重大修改
- 次版本号：新增功能或重要改进
- 修订号：Bug修复和文档更新

**合规性要求**：
- 所有代码审查必须验证是否符合宪法原则
- 项目复杂性增加时必须提供充分的理由说明
- 定期进行宪法合规性审查和更新

**决策机制**：
- 涉及架构原则的修改需要团队核心成员一致同意
- 日常开发中的原则解释由架构师负责
- 争议问题通过团队讨论解决

**版本**: 1.2.0 | **制定**: 2025-11-06 | **最后修订**: 2025-12-21
