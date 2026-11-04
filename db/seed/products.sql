INSERT INTO products (
    name,
    sku,
    description,
    price,
    stock_quantity,
    status
)
VALUES
    (
        'Wireless Mouse',
        'MOUSE-001',
        'Ergonomic wireless computer mouse',
        25.50,
        30,
        'active'
    ),
    (
        'Mechanical Keyboard',
        'KEYBOARD-001',
        'RGB mechanical gaming keyboard',
        75.00,
        15,
        'active'
    ),
    (
        'USB-C Hub',
        'HUB-001',
        'Seven-in-one USB-C adapter',
        45.99,
        20,
        'active'
    ),
    (
        'Laptop Stand',
        'STAND-001',
        'Adjustable aluminum laptop stand',
        35.00,
        12,
        'active'
    ),
    (
        'Webcam',
        'WEBCAM-001',
        'Full HD 1080p USB webcam',
        55.00,
        8,
        'active'
    ),
    (
        'Headphones',
        'HEADPHONE-001',
        'Over-ear wireless headphones',
        89.99,
        18,
        'active'
    ),
    (
        'External SSD',
        'SSD-001',
        'Portable 1TB external solid-state drive',
        120.00,
        10,
        'active'
    ),
    (
        'USB Flash Drive',
        'USB-001',
        '64GB USB 3.0 flash drive',
        12.50,
        50,
        'active'
    ),
    (
        'Monitor',
        'MONITOR-001',
        '24-inch Full HD computer monitor',
        180.00,
        6,
        'active'
    ),
    (
        'Laptop Backpack',
        'BAG-001',
        'Water-resistant laptop backpack',
        40.00,
        0,
        'draft'
    )
ON CONFLICT (sku) DO NOTHING;