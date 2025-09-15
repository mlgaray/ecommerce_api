DO $$
    DECLARE
userId bigint;


BEGIN
SELECT id INTO userId FROM users WHERE email = 'r.macdonalds@mc.com' LIMIT 1;


DELETE FROM user_roles WHERE user_id IN (userId);
END $$;