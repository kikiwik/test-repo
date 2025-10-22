# Personal Task API

这是一个基于 Go (Gin, GORM) 构建的个人任务管理 API。

## 技术栈

  * **后端:** Go 1.21
  * **Web 框架:** Gin
  * **ORM:** GORM
  * **数据库:** MySQL 8.0
  * **认证:** JWT
  * **部署:** Docker & Docker Compose

## 核心功能

  * **用户认证:** 基于 JWT 的注册与登录。
  * **任务 (Task) 管理:** 对任务的增、删、改、查 (CRUD)。
  * **项目 (Project) 管理:** 对项目的 CRUD。
  * **分类 (Category) 管理:** 对分类的 CRUD。
  * [cite\_start]**统计分析:** 提供任务概览、效率、每日/每周/每月报告等统计接口 [cite: 12]。
  * **权限控制:** 中间件确保用户只能访问自己的资源。

## 快速启动 (使用 Docker)

这是推荐的运行方式。

1.  **克隆项目**

    ```bash
    git clone <your-repo-url>
    cd test-repo
    ```

2.  **配置 (可选)**
    所有环境变量已在 `docker-compose.yml` 文件中硬编码。如果需要修改数据库密码或 JWT 密钥，请直接编辑此文件。

3.  **启动服务**
    此命令将同时构建 Go 应用镜像并启动 `app` 和 `db` (MySQL) 两个容器。

    ```bash
    docker-compose up -d --build
    ```

4.  **访问**

      * **API 服务:** `http://localhost:8080`
      * **数据库 (MySQL):** `localhost:3306`

## 本地开发 (不使用 Docker)

1.  **配置环境变量**
    复制 `.env` 文件 并根据需要修改。

    ```bash
    # 示例 .env
    ENVIRONMENT=development
    SERVER_PORT=8080
    DB_HOST=localhost
    DB_PORT=3306
    DB_USER=root
    DB_PASSWORD=114514
    DB_NAME=taskmanagement
    JWT_SECRET=your-super-secret-key
    ```

    *确保你本地有一个正在运行的 MySQL 实例，且配置与 `.env` 一致。*

2.  **安装依赖**

    ```bash
    go mod tidy
    ```

3.  **运行**

    ```bash
    go run main.go
    ```

    [cite\_start]服务器将在 `http://localhost:8080` 启动 [cite: 1]。

## 环境变量

应用通过 `config/config.go` 从环境变量加载配置：

| 变量名 | 描述 | `docker-compose.yml` 中的默认值 |
| :--- | :--- | :--- |
| `ENVIRONMENT` | 运行环境 (`development` 或 `production`) | `development` |
| `SERVER_PORT` | API 服务器端口 | `8080` |
| `DB_HOST` | 数据库主机 | `db` |
| `DB_PORT` | 数据库端口 | `3306` |
| `DB_USER` | 数据库用户名 | `root` |
| `DB_PASSWORD` | 数据库密码 | `114514` |
| `DB_NAME` | 数据库名称 | `taskmanagement` |
| `JWT_SECRET` | JWT 签名密钥 | `your-super-secret-key` |

-----

## API 端点

所有端点均以 `/api` 为前缀。

### 健康检查

  * `GET /health`: 检查 API 运行状态。

### 认证 `/api/auth`

  * `POST /register`: 用户注册。
  * `POST /login`: 用户登录，返回 JWT。
  * `GET /profile` (需认证): 获取当前用户信息。
  * `PUT /profile` (需认证): 更新当前用户信息。

### 任务 `/api/tasks` (需认证)

  * `GET /`: 获取任务列表（支持分页、过滤、排序）。
  * `POST /`: 创建新任务。
  * `GET /:id`: 获取单个任务详情。
  * `PUT /:id`: 更新任务。
  * `DELETE /:id`: 删除任务。
  * `PATCH /:id/status`: 更新任务状态。
  * `PATCH /batch/status`: 批量更新任务状态。
  * `DELETE /batch`: 批量删除任务。

### 分类 `/api/categories` (需认证)

  * `GET /`: 获取分类列表。
  * `POST /`: 创建新分类。
  * `GET /:id`: 获取分类详情（支持 `?with_tasks=true`）。
  * `PUT /:id`: 更新分类。
  * `DELETE /:id`: 删除分类（支持 `?force=true` 强制删除）。
  * `GET /:id/stats`: 获取分类统计。

### 项目 `/api/projects` (需认证)

  * `GET /`: 获取项目列表（支持 `?with_stats=true`）。
  * `POST /`: 创建新项目。
  * `GET /:id`: 获取项目详情（支持 `?with_tasks=true`）。
  * `PUT /:id`: 更新项目。
  * `DELETE /:id`: 删除项目（支持 `?force=true` 强制删除）。
  * `GET /:id/tasks`: 获取项目下的任务列表。
  * `GET /:id/stats`: 获取项目统计。

### 统计 `/api/stats` (需认证)

  * `GET /overview`: 任务概览统计。
  * [cite\_start]`GET /daily`: 每日任务统计（支持 `?days=N`）[cite: 12]。
  * [cite\_start]`GET /weekly`: 每周任务统计（支持 `?weeks=N`）[cite: 12]。
  * `GET /productivity`: 工作效率分析。
  * [cite\_start]`GET /monthly`: 月度报告（支持 `?month=YYYY-MM`）[cite: 12]。
