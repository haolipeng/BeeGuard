-- 000009: 回滚系统管理相关表

DROP TRIGGER IF EXISTS trigger_intrusion_rules_updated_at ON hids_rules;
DROP TABLE IF EXISTS hids_rules;
DROP TABLE IF EXISTS systen_user;
