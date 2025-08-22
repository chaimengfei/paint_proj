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