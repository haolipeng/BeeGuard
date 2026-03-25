-- 000005 down: 删除基线表（按依赖逆序）
DROP TABLE IF EXISTS baseline_check_detail CASCADE;
DROP TABLE IF EXISTS baseline_check_result CASCADE;
DROP TABLE IF EXISTS baseline_check_item CASCADE;
DROP TABLE IF EXISTS baseline_template_host_link CASCADE;
DROP TABLE IF EXISTS baseline_template CASCADE;
