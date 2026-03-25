-- =====================================================
-- 模拟数据: image_vuln_scan_task (镜像漏洞扫描任务表)
-- 数据量: 15条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
--       VPC CIDR: 10.0.0.0/16
--       引用 012_mock_asset_image.sql 中的镜像数据
--       scan_status: 1-成功
--       scan_trigger: auto-自动扫描
--       matched_vulns 与 image_vuln_detail 数据匹配
-- =====================================================

INSERT INTO image_vuln_scan_task (id, agent_id, image_id, image_name, scan_status, scan_trigger, total_packages, matched_vulns, scan_duration, error_message, scan_time, created_at, updated_at) VALUES
-- goharbor/nginx-photon 镜像 (aws-harbor-01)
(1,  'agent-030-m7n8o9p0', 'sha256:a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8',
 'goharbor/nginx-photon:v2.10.0',                1, 'auto', 87,  5, 2150, NULL, NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),
-- redis 镜像 (aws-eks-node-02)
(2,  'agent-025-s7t8u9v0', 'sha256:01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2',
 'redis:7.2-alpine',                             1, 'auto', 42,  3, 980,  NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '55 days', NOW()),
-- mysql 镜像 (aws-eks-node-03)
(3,  'agent-026-w1x2y3z4', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4',
 'mysql:8.0',                                    1, 'auto', 138, 4, 3250, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '57 days', NOW()),
-- postgres 镜像 (aws-jenkins-01)
(4,  'agent-028-e9f0g1h2', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4',
 'postgres:16-alpine',                           1, 'auto', 48,  4, 1120, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),
-- elasticsearch 镜像 (aws-eks-node-04)
(5,  'agent-027-a5b6c7d8', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 1, 'auto', 112, 4, 2870, NULL, NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '55 days', NOW()),
-- company/frontend 镜像 (aws-eks-node-01)
(6,  'agent-024-o3p4q5r6', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8',
 '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend:v3.2.0', 1, 'auto', 95, 4, 2340, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
-- company/backend 镜像 (aws-eks-node-01)
(7,  'agent-024-o3p4q5r6', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9',
 '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend:v4.1.0', 1, 'auto', 103, 4, 2560, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
-- ingress-nginx 镜像 (aws-eks-node-01)
(8,  'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 1, 'auto', 78, 3, 1850, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
-- harbor-core 镜像 (aws-harbor-01)
(9,  'agent-030-m7n8o9p0', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
 'goharbor/harbor-core:v2.10.0',                 1, 'auto', 125, 3, 3080, NULL, NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '80 days', NOW()),
-- gitlab-runner 镜像 (aws-gitlab-01)
(10, 'agent-029-i3j4k5l6', 'sha256:90d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1',
 'gitlab/gitlab-runner:v16.9.1',                 1, 'auto', 89,  3, 2190, NULL, NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '95 days', NOW()),
-- golang 镜像 (aws-gitlab-01)
(11, 'agent-029-i3j4k5l6', 'sha256:34b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5',
 'golang:1.22-alpine',                           1, 'auto', 56,  3, 1350, NULL, NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '85 days', NOW()),
-- prom/alertmanager 镜像 (aws-prometheus-01)
(12, 'agent-033-y9z0a1b2', 'sha256:78f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9',
 'prom/alertmanager:v0.27.0',                    1, 'auto', 67,  2, 1580, NULL, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '100 days', NOW()),
-- company/worker 镜像 (aws-eks-node-03)
(13, 'agent-026-w1x2y3z4', 'sha256:34d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5',
 '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/worker:v2.3.0', 1, 'auto', 98, 3, 2410, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '50 days', NOW()),
-- python 镜像 (aws-gitlab-01)
(14, 'agent-029-i3j4k5l6', 'sha256:45c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6',
 'python:3.12-slim',                             1, 'auto', 142, 3, 3420, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '20 days', NOW()),
-- node 镜像 (aws-gitlab-01)
(15, 'agent-029-i3j4k5l6', 'sha256:23a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4',
 'node:20-alpine',                               1, 'auto', 74,  2, 1720, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '15 days', NOW());

-- 重置序列
SELECT setval('image_vuln_scan_task_id_seq', 15);
