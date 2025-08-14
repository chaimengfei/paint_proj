# 油漆销售管理系统

## 功能特性

### 关于登陆

1. 是否每次都要调用 GetOpenIDByCode?
**不是每次都要调用。标准流程如下：**


| 阶段             | 操作                                       | 调用 GetOpenIDByCode？ | 后端是否需要用它？ |
| ---------------- | ------------------------------------------ | ---------------------- | ------------------ |
| 第一次登录小程序 | wx.login() + 后端换取 openid + 绑定 userID | ✅ 是                   | ✅ 是               |
| 后续所有请求     | 带上 token 或 userID                       | ❌ 否                   | ❌ 不需要调用       |
| 发起支付         | 用存下来的 openid 发起支付下单             | ❌ 否                   | ✅ 用 openid        |

2. 微信小程序中，`openid` 可以作为用户唯一标识，是否可以替代 `user_id`？

| 方案                         | 是否可行 | 说明                                                         |
| ---------------------------- | -------- | ------------------------------------------------------------ |
| 直接用 `openid` 作为用户主键 | ✅ 可行   | 若你项目只对接微信小程序，openid 是稳定主键                  |
| 使用 `user_id`（数字ID）     | ✅ 推荐   | 更通用、便于关联数据库中的其他表（如订单、评论、积分等）     |
| 两者都存                     | ✅ 最推荐 | 保留 `user_id`（主键自增ID），记录 `openid`（平台标识），便于扩展和对接其他渠道（如 Web、App、小程序等） |

 ✅ 使用方式推荐

| 场景             | 建议使用                                                     |
| ---------------- | ------------------------------------------------------------ |
| 用户注册、登录时 | 使用 `openid` 查找/创建用户                                  |
| 数据库关联、查询 | 使用 `user_id`                                               |
| 接口认证         | 生成 Token 时包含 `user_id` 或 `openid`，推荐以 `user_id` 为主 |


### 库存管理功能

#### 1. 小程序用户购买自动出库
- 当用户在小程序中下单时，系统会自动：
  - 检查商品库存是否充足
  - 创建订单记录
  - 自动减少商品库存
  - 记录库存出库日志

#### 2. 管理员后台库存操作
- **入库操作** (`POST /admin/stock/inbound`)
  - 管理员可以手动增加商品库存
  - 记录入库日志
  
- **出库操作** (`POST /admin/stock/outbound`)
  - 管理员可以手动减少商品库存
  - 检查库存是否充足
  - 记录出库日志
  
- **退货操作** (`POST /admin/stock/return`)
  - 管理员可以处理退货，增加商品库存
  - 记录退货日志

#### 3. 库存日志查询
- **获取库存日志** (`GET /admin/stock/logs`)
  - 支持按商品ID筛选
  - 支持分页查询
  - 记录所有库存操作历史

### API接口说明

#### 库存管理接口

##### 入库操作
```http
POST /admin/stock/inbound
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 10,
  "remark": "新货入库"
}
```

##### 出库操作
```http
POST /admin/stock/outbound
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 5,
  "remark": "手动出库"
}
```

##### 退货操作
```http
POST /admin/stock/return
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 2,
  "remark": "客户退货"
}
```

##### 获取库存日志
```http
GET /admin/stock/logs?product_id=1&page=1&page_size=10
```

### 数据库表结构

#### stock_log 库存日志表
```sql
CREATE TABLE stock_log (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  product_id BIGINT NOT NULL COMMENT '商品ID',
  product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
  type TINYINT NOT NULL COMMENT '操作类型(1:入库,2:出库,3:退货)',
  quantity INT NOT NULL COMMENT '操作数量',
  before_stock INT NOT NULL COMMENT '操作前库存',
  after_stock INT NOT NULL COMMENT '操作后库存',
  order_no VARCHAR(64) COMMENT '关联订单号(出库/退货时)',
  remark VARCHAR(255) COMMENT '备注',
  operator VARCHAR(255) NOT NULL COMMENT '操作人',
  operator_type TINYINT NOT NULL COMMENT '操作人类型(1:用户,2:系统,3:管理员)',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
```

### 使用说明

1. **小程序购买流程**：
   - 用户选择商品并下单
   - 系统自动检查库存
   - 创建订单并减少库存
   - 记录出库日志

2. **管理员库存管理**：
   - 通过后台管理界面进行入库、出库、退货操作
   - 所有操作都会记录详细的日志
   - 支持库存日志查询和追溯

3. **库存安全**：
   - 所有库存操作都有库存检查
   - 出库时会验证库存是否充足
   - 完整的操作日志记录