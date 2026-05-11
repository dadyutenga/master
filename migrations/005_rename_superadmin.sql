-- 005_rename_superadmin.sql
UPDATE users SET role = 'admin' WHERE role = 'superadmin';
