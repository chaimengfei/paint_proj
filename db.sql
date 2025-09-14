-- 创建库存日志表
CREATE TABLE IF NOT EXISTS stock_log (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键id',
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
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  INDEX idx_product_id (product_id),
  INDEX idx_type (type),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存日志表';

ALTER TABLE product ADD COLUMN is_on_shelf TINYINT NOT NULL DEFAULT 0 COMMENT '是否上架(1:上架,0:下架)';

CREATE TABLE stock_operation (
     id BIGINT PRIMARY KEY AUTO_INCREMENT,
     operation_no VARCHAR(64) NOT NULL COMMENT '操作单号',
     type TINYINT NOT NULL COMMENT '操作类型(1:入库,2:出库,3:退货)',
     operator VARCHAR(255) NOT NULL COMMENT '操作人',
     operator_id BIGINT NOT NULL COMMENT '操作人ID',
     operator_type TINYINT NOT NULL COMMENT '操作人类型(1:用户,2:系统,3:管理员)',
     user_name VARCHAR(255) COMMENT '用户名称(出库时)',
     user_id BIGINT COMMENT '用户ID(出库时)',
     user_account VARCHAR(255) COMMENT '用户账号(出库时)',
     remark VARCHAR(255) COMMENT '备注',
     total_amount BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
) COMMENT '库存操作主表';


CREATE TABLE IF NOT EXISTS stock_operation_item (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键id',
    operation_id BIGINT NOT NULL COMMENT '操作主表ID',
    product_id BIGINT NOT NULL COMMENT '商品ID',
    product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
    quantity INT NOT NULL COMMENT '操作数量',
    unit_price BIGINT NOT NULL DEFAULT 0 COMMENT '单价(分)',
    total_price BIGINT NOT NULL DEFAULT 0 COMMENT '总价(分)',
    before_stock INT NOT NULL COMMENT '操作前库存',
    after_stock INT NOT NULL COMMENT '操作后库存',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_operation_id (operation_id),
    INDEX idx_product_id (product_id),
    FOREIGN KEY (operation_id) REFERENCES stock_operation(id) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存操作子表';

ALTER TABLE stock_operation
CREATE INDEX idx_user_id ON stock_operation(user_id);


-- 为product表添加specification字段
ALTER TABLE product ADD COLUMN specification VARCHAR(200) DEFAULT '' COMMENT '规格';

-- 为stock_operation_item表添加specification字段
ALTER TABLE stock_operation_item ADD COLUMN specification VARCHAR(200) DEFAULT '' COMMENT '规格';

-- 为stock_operation_item表添加remark字段
ALTER TABLE stock_operation_item ADD COLUMN remark VARCHAR(500) DEFAULT '' COMMENT '备注';


-- 创建入库成本变更记录表
CREATE TABLE inbound_cost_change (
         id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键id',
         operation_id BIGINT NOT NULL COMMENT '入库操作ID',
         product_id BIGINT NOT NULL COMMENT '商品ID',
         product_name VARCHAR(255) NOT NULL COMMENT '商品名称',
         old_cost BIGINT NOT NULL DEFAULT 0 COMMENT '原成本价(分)',
         new_cost BIGINT NOT NULL DEFAULT 0 COMMENT '新成本价(分)',
         change_reason VARCHAR(500) DEFAULT '' COMMENT '变更原因',
         operator VARCHAR(100) NOT NULL COMMENT '操作人',
         operator_id BIGINT NOT NULL COMMENT '操作人ID',
         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
         INDEX idx_operation_id (operation_id),
         INDEX idx_product_id (product_id),
         INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='入库成本变更记录表';



-- 为stock_operation表添加outbound_type字段
ALTER TABLE stock_operation
    ADD COLUMN outbound_type TINYINT DEFAULT NULL COMMENT '出库类型(1:小程序购买,2:admin后台操作)' AFTER types;

-- 为stock_operation_item表添加成本相关字段
ALTER TABLE stock_operation_item
    ADD COLUMN cost BIGINT NOT NULL DEFAULT 0 COMMENT '成本价(暂不用) 单位:分' AFTER after_stock,
    ADD COLUMN shipping_cost BIGINT NOT NULL DEFAULT 0 COMMENT '运费(暂不用) 单位:分' AFTER cost,
    ADD COLUMN product_cost BIGINT NOT NULL DEFAULT 0 COMMENT '货物成本(暂不用) 单位:分' AFTER shipping_cost;

-- 为stock_operation_item表添加订单关联字段
ALTER TABLE stock_operation_item
    ADD COLUMN order_id BIGINT DEFAULT NULL COMMENT '关联订单ID(小程序购买时)' AFTER operation_id,
    ADD COLUMN order_no VARCHAR(64) DEFAULT NULL COMMENT '关联订单号(小程序购买时)' AFTER order_id,
    ADD INDEX idx_order_id (order_id),
    ADD INDEX idx_order_no (order_no);

-- 为order_log表添加operator_id字段，与stock_operation表保持一致
ALTER TABLE order_log
    ADD COLUMN operator_id BIGINT DEFAULT NULL COMMENT '操作人ID' AFTER operator,
    ADD INDEX idx_operator_id (operator_id);

-- 为stock_operation_item表添加unit字段
ALTER TABLE stock_operation_item
    ADD COLUMN unit VARCHAR(100) NOT NULL DEFAULT '' COMMENT '单位 L/桶/套' AFTER specification;

-- 为stock_operation_item表添加利润相关字段
ALTER TABLE stock_operation_item
    ADD COLUMN profit BIGINT NOT NULL DEFAULT 0 COMMENT '利润(卖价-成本价)*数量 单位:分' AFTER product_cost;

-- 为stock_operation表添加总利润字段
ALTER TABLE stock_operation
    ADD COLUMN total_profit BIGINT NOT NULL DEFAULT 0 COMMENT '总利润 单位:分' AFTER total_amount;

-- 为stock_operation表添加支付完成状态字段
ALTER TABLE stock_operation
    ADD COLUMN payment_finish_status TINYINT NOT NULL DEFAULT 1 COMMENT '支付完成状态(1:未支付,3:已支付)' AFTER total_profit,
    ADD COLUMN payment_finish_time TIMESTAMP NULL COMMENT '支付完成时间' AFTER payment_finish_status;

-- 为stock_operation表添加供货商字段
ALTER TABLE stock_operation
    ADD COLUMN supplier VARCHAR(255) DEFAULT '' COMMENT '供货商' AFTER payment_finish_time;

-- 创建供货商表
CREATE TABLE IF NOT EXISTS supplier (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '供货商ID',
    name VARCHAR(500) NOT NULL COMMENT '供货商名称',
    area VARCHAR(255) DEFAULT '' COMMENT '供货商所在地区',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_name (name),
    INDEX idx_area (area)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供货商表';

-- 为stock_operation表添加总数量字段
ALTER TABLE stock_operation
    ADD COLUMN total_quantity INT NOT NULL DEFAULT 0 COMMENT '总数量' AFTER total_amount;

-- 为用户表添加新字段，支持小程序和后台管理系统共用
ALTER TABLE user 
    ADD COLUMN mobile_phone VARCHAR(20) UNIQUE COMMENT '手机号（唯一标识）' AFTER avatar,
    ADD COLUMN source TINYINT NOT NULL DEFAULT 1 COMMENT '用户来源(1:小程序,2:后台添加,3:混合)' AFTER mobile_phone,
    ADD COLUMN is_enable TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用(1:启用,0:禁用)' AFTER source,
    ADD COLUMN admin_display_name VARCHAR(100) COMMENT '后台管理系统显示的客户名称' AFTER is_enable,
    ADD COLUMN wechat_display_name VARCHAR(100) COMMENT '微信小程序显示的客户名称' AFTER admin_display_name,
    ADD COLUMN has_wechat_bind TINYINT NOT NULL DEFAULT 0 COMMENT '是否已绑定微信(1:是,0:否)' AFTER wechat_display_name;

-- 为现有用户数据设置默认值
UPDATE user SET 
    source = 1,
    has_wechat_bind = CASE WHEN openid IS NOT NULL AND openid != '' THEN 1 ELSE 0 END,
    wechat_display_name = CASE WHEN nickname IS NOT NULL AND nickname != '' THEN nickname ELSE '' END
WHERE id > 0;