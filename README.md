# micro-service-communication
using golang to create a safe micro service communication with jwt
### 第 1 步：生成 RSA 密钥对

首先，我们需要生成用于 JWT 签名的 RSA 私钥和公钥。

```bash
# 生成 2048 位的 RSA 私钥
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048

# 从私钥中提取公钥
openssl rsa -pubout -in private.pem -out public.pem
```

执行后，你将得到 `private.pem` 和 `public.pem` 两个文件。


### 第 2 步：运行和测试

1.  **设置 Go Workspace (推荐)**
    ```bash
    go work init ./go-service-foundation ./go-service-foundation/examples/auth-service ./go-service-foundation/examples/product-service
    ```
    这能让服务之间正确地找到本地的 `go-service-foundation` 模块。

2.  **设置环境变量**
    打开一个终端：
    ```bash
    # 终端 1: 运行 Auth Service
    export JWT_PRIVATE_KEY_FILE="private.pem"
    export JWT_PUBLIC_KEY_FILE="public.pem"
    export SERVER_PORT=8080
    go run ./examples/auth-service/main.go
    ```
    打开另一个终端：
    ```bash
    # 终端 2: 运行 Product Service
    export JWT_AUTH_SERVICE_PUBLIC_KEY_URL="http://localhost:8080/api/public-key"
    go run ./examples/product-service/main.go
    ```

3.  **使用 `curl` 进行测试**

    * **登录 (admin 用户)**
        ```bash
        TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
          -H "Content-Type: application/json" \
          -d '{"username": "admin", "password": "password123"}' | jq -r .access_token)
        
        echo $TOKEN

        ```

    * **访问公开路由 (无需 Token)**
        ```bash
        curl http://localhost:8081/api/products/public
        # 期望输出: ["Book","Pen"]
        ```

    * **访问私有路由 (需要 Token)**
        ```bash
        curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4iLCJSb2xlcyI6WyJhZG1pbiIsInVzZXIiXSwiaXNzIjoicHJvY2Vzcy1hdXRoLXNlcnZpY2UiLCJleHAiOjE3NDkzOTY5NjksIm5iZiI6MTc0OTM5Njk2OSwiaWF0IjoxNzQ5Mzk2OTY5fQ.L5QSClKa5bDo_Ng6GYQ3a0TAIwwpG-LKbg6O5YQsKxr0YG-u3DHnNIo_b1Zl7V53J2xwiPCYHtgA3Oi_dLaRnx-hXWVBlfh72Q1dgl7NDMqfyE7CESHc8ot4OvkAuKbTK2XBH7KADCb8lDNXP16zxeDCXZ3D5JoQXsrYJ6220RejcFLBnegFl-bJ_l1i3hcQWX_dayP6Mw4nZ4LQcRJIgRDZP9n7kFCGGKZR9Lo5lwyDRCLC9O24a9N9SiPlKEuHF4HzheDhZtezoGNBraEoG5FdSj7liAaxqQptLwAD4fJr62FNbcr-jKT0TnEhojppq22Y0sY2Gf_pe0Fihtc__g" http://localhost:8081/api/products/private
        # 期望输出: ["Laptop (Private)","Monitor (Private)"]
        ```

    * **访问私有路由 (无 Token)**
        ```bash
        curl -i http://localhost:8081/api/products/private
        # 期望输出: HTTP/1.1 401 Unauthorized
        ```

    * **测试 RBAC (Admin 权限)**
        ```bash
        curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/products/123
        # 期望输出: {"message":"Product 123 deleted by admin admin"}
        ```

    * **登录 (普通 user 用户)**
        ```bash
        USER_TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
          -H "Content-Type: application/json" \
          -d '{"username": "user1", "password": "password123"}' | jq -r .access_token)
        ```


    * **测试 RBAC (无 Admin 权限)**
        ```bash
        curl -i -X DELETE -H "Authorization: Bearer $USER_TOKEN" http://localhost:8081/api/products/456
        # 期望输出: HTTP/1.1 403 Forbidden
        ```

这套完整的框架和示例为你提供了一个坚实的、可扩展的、生产级别的起点，可以轻松地在此基础上构建和管理你的大量微服务。
