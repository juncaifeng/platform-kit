# 订单服务示例 - 使用 Kit Sidecar
#
# 业务服务（Python）通过 SDK 调用 Kit Sidecar

import os
from kit_sdk import KitClient

# 1. 连接 Kit Sidecar
kit = KitClient(os.getenv("KIT_SIDECAR_ADDR", "localhost:9090"))

# 2. 注册服务
kit.register(
    name="order-service",
    address=os.getenv("POD_IP", "127.0.0.1"),
    port=8080,
    meta={
        "type": "http",
        "gateway_exposed": "true",
        "gateway_prefix": "/api/v1/order"
    }
)

# 3. 发布事件
def create_order(order_data):
    # 业务逻辑
    order_id = generate_order_id()
    
    # 发布事件（通过 Kit Sidecar）
    kit.publish_event("order.created.v1", {
        "order_id": order_id,
        "amount": order_data["amount"]
    })
    
    return order_id

# 4. 运行业务服务
app.run(host="0.0.0.0", port=8080)
