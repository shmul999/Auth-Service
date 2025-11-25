INSERT INTO apps (id, name, secret)
SELECT 1, 'test', 'test-secret'
WHERE NOT EXISTS (SELECT 1 FROM apps WHERE id = 1 OR name = 'test');