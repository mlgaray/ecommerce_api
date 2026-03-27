INSERT INTO users (name, last_name, email, phone, password, is_active)
VALUES ('Ronald', 'McDonald´s', 'r.macdonalds@mc.com', '3512658244', crypt('As123456!', gen_salt('bf')), true)
