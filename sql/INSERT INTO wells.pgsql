INSERT INTO wells 
(name, depth, location, status, productivity, drilling_date, field, operator)
VALUES 
('Test Well', 1000, 'Test Location', 'active', 100, '2020-01-01', 'Test Field', 'Test Operator')
RETURNING id;