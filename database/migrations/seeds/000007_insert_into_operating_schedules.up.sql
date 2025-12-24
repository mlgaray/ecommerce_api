-- Operating schedules for test shop (shop_id = 1)
-- Monday to Friday: 9:00-13:00 and 17:00-21:00 (split hours)
-- Saturday: 10:00-14:00 (single shift)
-- Sunday: closed (no rows)

INSERT INTO operating_schedules (shop_id, day_of_week, open_time, close_time) VALUES
-- Monday (1)
(1, 1, '09:00', '13:00'),
(1, 1, '17:00', '21:00'),
-- Tuesday (2)
(1, 2, '09:00', '13:00'),
(1, 2, '17:00', '21:00'),
-- Wednesday (3)
(1, 3, '09:00', '13:00'),
(1, 3, '17:00', '21:00'),
-- Thursday (4)
(1, 4, '09:00', '13:00'),
(1, 4, '17:00', '21:00'),
-- Friday (5)
(1, 5, '09:00', '13:00'),
(1, 5, '17:00', '21:00'),
-- Saturday (6)
(1, 6, '10:00', '14:00');
-- Sunday (0) - closed, no rows
