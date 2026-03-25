-- 000006 down: 删除漏洞表（按依赖逆序）
DROP TABLE IF EXISTS image_vuln_detail CASCADE;
DROP TABLE IF EXISTS image_vuln_scan_task CASCADE;
DROP TABLE IF EXISTS host_vuln_detail CASCADE;
DROP TABLE IF EXISTS host_vuln_scan_task CASCADE;
DROP TABLE IF EXISTS vuln_info CASCADE;
DROP TABLE IF EXISTS image_vulnerability_info CASCADE;
DROP TABLE IF EXISTS vulnerability_info CASCADE;
