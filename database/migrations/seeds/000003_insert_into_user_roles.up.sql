DO $$
    DECLARE
userId bigint;
roleId bigint;

BEGIN
SELECT id INTO userId FROM users WHERE email = 'r.macdonalds@mc.com' LIMIT 1;
SELECT id INTO roleId FROM roles WHERE name = 'admin' LIMIT 1;

INSERT INTO user_roles (user_id , role_id)
VALUES (userId,roleId);
END $$;