-- =====================================================
-- 模拟数据: asset_image (镜像资产表)
-- 数据量: 80条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的主机生成镜像数据
-- 镜像来源: Harbor, Jenkins CI, EKS/K8s, GitLab Runner, 监控
-- =====================================================

INSERT INTO asset_image (agent_id, host_name, host_ip, image_id, image_name, image_version, image_size, container_count, build_time, created_at, updated_at) VALUES

-- ==========================================
-- aws-harbor-01 镜像 (Harbor 服务组件)
-- agent-030-m7n8o9p0 / 10.0.5.12
-- ==========================================
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2', 'goharbor/harbor-core', 'v2.10.0', 148897792, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3', 'goharbor/harbor-portal', 'v2.10.0', 61865984, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4', 'goharbor/harbor-db', 'v2.10.0', 251658240, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5', 'goharbor/redis-photon', 'v2.10.0', 130023424, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6', 'goharbor/registry-photon', 'v2.10.0', 92274688, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7', 'goharbor/harbor-jobservice', 'v2.10.0', 141557760, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8', 'goharbor/nginx-photon', 'v2.10.0', 48234496, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9', 'goharbor/harbor-registryctl', 'v2.10.0', 136314880, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0', 'goharbor/trivy-adapter-photon', 'v2.10.0', 167772160, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),

-- ==========================================
-- aws-jenkins-01 镜像 (CI/CD 构建镜像)
-- agent-028-e9f0g1h2 / 10.0.5.10
-- ==========================================
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1', 'jenkins/jenkins', '2.440-lts-jdk17', 471859200, 1, '2024-02-10 08:00:00', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2', 'jenkins/inbound-agent', '3206.vb_15dcf73f6a_9-3-jdk17', 378535936, 2, '2024-02-10 10:00:00', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3', 'sonarqube', '10.4-community', 603979776, 1, '2024-01-20 14:00:00', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4', 'postgres', '16-alpine', 247463936, 1, '2024-01-20 14:00:00', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5', 'maven', '3.9-eclipse-temurin-21', 545259520, 0, '2024-03-01 15:00:00', NOW() - INTERVAL '5 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6', 'docker', '25.0-dind', 314572800, 1, '2024-02-15 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/build-base', 'v1.4.0', 628097024, 1, '2024-02-20 11:00:00', NOW() - INTERVAL '30 days', NOW()),

-- ==========================================
-- aws-eks-master-01 镜像 (K8s 控制面组件)
-- agent-023-k9l0m1n2 / 10.0.4.10
-- ==========================================
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8', 'registry.k8s.io/etcd', '3.5.12-0', 152043520, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9', 'registry.k8s.io/kube-apiserver', 'v1.29.2', 130023424, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0', 'registry.k8s.io/kube-controller-manager', 'v1.29.2', 126877696, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1', 'registry.k8s.io/kube-scheduler', 'v1.29.2', 62914560, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2', 'registry.k8s.io/coredns/coredns', 'v1.11.1', 53477376, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3', 'docker.io/calico/kube-controllers', 'v3.27.2', 73400320, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'sha256:e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4', 'registry.k8s.io/kube-proxy', 'v1.29.2', 75497472, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '60 days', NOW()),

-- ==========================================
-- aws-eks-node-01 镜像 (K8s 工作节点 1)
-- agent-024-o3p4q5r6 / 10.0.4.11
-- ==========================================
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4', 'registry.k8s.io/kube-proxy', 'v1.29.2', 75497472, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5', 'docker.io/calico/node', 'v3.27.2', 215482368, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6', 'docker.io/calico/cni', 'v3.27.2', 199229440, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7', 'registry.k8s.io/ingress-nginx/controller', 'v1.10.0', 299892736, 1, '2024-02-01 09:00:00', NOW() - INTERVAL '55 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend', 'v3.2.0', 162529280, 2, '2024-03-05 11:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend', 'v4.1.0', 289406976, 2, '2024-03-05 11:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0', 'prom/node-exporter', 'v1.7.0', 22020096, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1', 'registry.k8s.io/pause', '3.9', 746496, 4, '2024-01-20 09:00:00', NOW() - INTERVAL '58 days', NOW()),

-- ==========================================
-- aws-eks-node-02 镜像 (K8s 工作节点 2)
-- agent-025-s7t8u9v0 / 10.0.4.12
-- ==========================================
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4', 'registry.k8s.io/kube-proxy', 'v1.29.2', 75497472, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5', 'docker.io/calico/node', 'v3.27.2', 215482368, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend', 'v3.2.0', 162529280, 2, '2024-03-05 11:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend', 'v4.1.0', 289406976, 2, '2024-03-05 11:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2', 'redis', '7.2-alpine', 42991616, 1, '2024-02-10 15:00:00', NOW() - INTERVAL '55 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:12b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/payment-svc', 'v2.0.1', 204472320, 1, '2024-03-08 14:00:00', NOW() - INTERVAL '45 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0', 'prom/node-exporter', 'v1.7.0', 22020096, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1', 'registry.k8s.io/pause', '3.9', 746496, 4, '2024-01-20 09:00:00', NOW() - INTERVAL '56 days', NOW()),

-- ==========================================
-- aws-eks-node-03 镜像 (K8s 工作节点 3)
-- agent-026-w1x2y3z4 / 10.0.4.13
-- ==========================================
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4', 'registry.k8s.io/kube-proxy', 'v1.29.2', 75497472, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5', 'docker.io/calico/node', 'v3.27.2', 215482368, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4', 'mysql', '8.0', 489684992, 1, '2024-02-05 08:00:00', NOW() - INTERVAL '52 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:34d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/worker', 'v2.3.0', 205520896, 1, '2024-03-05 11:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:45e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/scheduler', 'v1.5.0', 178257920, 1, '2024-03-08 10:00:00', NOW() - INTERVAL '48 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0', 'prom/node-exporter', 'v1.7.0', 22020096, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'sha256:f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1', 'registry.k8s.io/pause', '3.9', 746496, 3, '2024-01-20 09:00:00', NOW() - INTERVAL '54 days', NOW()),

-- ==========================================
-- aws-eks-node-04 镜像 (K8s 工作节点 4)
-- agent-027-a5b6c7d8 / 10.0.4.14
-- ==========================================
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4', 'registry.k8s.io/kube-proxy', 'v1.29.2', 75497472, 1, '2024-01-20 09:00:00', NOW() - INTERVAL '52 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5', 'docker.io/calico/node', 'v3.27.2', 215482368, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '52 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7', 'docker.elastic.co/elasticsearch/elasticsearch', '8.12.2', 812646400, 1, '2024-02-15 16:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:67a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/notification-svc', 'v1.2.0', 157286400, 1, '2024-03-10 14:00:00', NOW() - INTERVAL '42 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:78b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/auth-svc', 'v2.1.0', 183500800, 1, '2024-03-10 14:00:00', NOW() - INTERVAL '42 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:89c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backup-job', 'v1.4.0', 94371840, 1, '2024-03-15 02:00:00', NOW() - INTERVAL '3 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0', 'prom/node-exporter', 'v1.7.0', 22020096, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '52 days', NOW()),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'sha256:f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1', 'registry.k8s.io/pause', '3.9', 746496, 4, '2024-01-20 09:00:00', NOW() - INTERVAL '52 days', NOW()),

-- ==========================================
-- aws-gitlab-01 镜像 (GitLab Runner 镜像)
-- agent-029-i3j4k5l6 / 10.0.5.11
-- ==========================================
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:90d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1', 'gitlab/gitlab-runner', 'v16.9.1', 436207616, 2, '2024-02-01 10:00:00', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:01e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2', 'gitlab/gitlab-runner-helper', 'x86_64-v16.9.1', 62914560, 2, '2024-02-01 10:00:00', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:12f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3', 'docker', '25.0-dind', 314572800, 1, '2024-02-15 09:00:00', NOW() - INTERVAL '60 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:23a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4', 'node', '20-alpine', 183500800, 0, '2024-03-01 08:00:00', NOW() - INTERVAL '15 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:34b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5', 'golang', '1.22-alpine', 268435456, 0, '2024-02-20 08:00:00', NOW() - INTERVAL '30 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:45c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6', 'python', '3.12-slim', 162529280, 0, '2024-02-25 08:00:00', NOW() - INTERVAL '25 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'sha256:56d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/ci-base', 'v2.0.0', 524288000, 1, '2024-02-28 09:00:00', NOW() - INTERVAL '20 days', NOW()),

-- ==========================================
-- aws-prometheus-01 镜像 (Prometheus 监控栈)
-- agent-033-y9z0a1b2 / 10.0.6.10
-- ==========================================
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:67e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8', 'prom/prometheus', 'v2.50.1', 257949696, 1, '2024-02-20 10:00:00', NOW() - INTERVAL '30 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:78f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9', 'prom/alertmanager', 'v0.27.0', 69206016, 1, '2024-02-15 09:00:00', NOW() - INTERVAL '35 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:89a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0', 'prom/blackbox-exporter', 'v0.25.0', 26214400, 1, '2024-01-10 14:00:00', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:90b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1', 'prom/pushgateway', 'v1.7.0', 22544384, 1, '2024-01-10 14:00:00', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:01c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2', 'quay.io/thanos/thanos', 'v0.34.1', 141557760, 2, '2024-02-01 09:00:00', NOW() - INTERVAL '50 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:12d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3', 'grafana/grafana', '10.4.1', 436207616, 1, '2024-03-01 10:00:00', NOW() - INTERVAL '25 days', NOW()),

-- ==========================================
-- aws-elk-01 镜像 (ELK 日志栈)
-- agent-035-g7h8i9j0 / 10.0.6.12
-- ==========================================
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:23e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4', 'docker.elastic.co/elasticsearch/elasticsearch', '8.12.2', 812646400, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:34f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5', 'docker.elastic.co/kibana/kibana', '8.12.2', 1102053376, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:45a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6', 'docker.elastic.co/logstash/logstash', '8.12.2', 859832320, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:56b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7', 'docker.elastic.co/beats/filebeat', '8.12.2', 325058560, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:67c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8', 'docker.elastic.co/apm/apm-server', '8.12.2', 220200960, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 补充镜像 (各节点额外镜像)
-- ==========================================
-- aws-eks-node-01: 额外业务镜像
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:a1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/gateway-svc', 'v2.0.3', 194641920, 1, '2024-03-08 11:00:00', NOW() - INTERVAL '48 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'sha256:b2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3', 'docker.io/calico/csi', 'v3.27.2', 20971520, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '58 days', NOW()),
-- aws-eks-node-02: 额外业务镜像
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:c3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4', '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/order-svc', 'v3.0.2', 231735296, 1, '2024-03-10 09:00:00', NOW() - INTERVAL '40 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'sha256:d4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5', 'docker.io/calico/csi', 'v3.27.2', 20971520, 1, '2024-01-18 09:00:00', NOW() - INTERVAL '56 days', NOW()),
-- aws-harbor-01: 额外扫描镜像
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'sha256:e5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6', 'goharbor/harbor-exporter', 'v2.10.0', 89128960, 1, '2024-01-15 08:00:00', NOW() - INTERVAL '80 days', NOW()),
-- aws-jenkins-01: 额外构建工具镜像
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'sha256:f6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7', 'gradle', '8.5-jdk17', 587202560, 0, '2024-02-28 08:00:00', NOW() - INTERVAL '20 days', NOW()),
-- aws-prometheus-01: 额外监控镜像
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'sha256:a7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8', 'jimmidyson/configmap-reload', 'v0.12.0', 12582912, 2, '2024-01-10 14:00:00', NOW() - INTERVAL '110 days', NOW()),
-- aws-elk-01: 额外日志采集镜像
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'sha256:b8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9', 'docker.elastic.co/beats/metricbeat', '8.12.2', 335544320, 1, '2024-02-15 08:00:00', NOW() - INTERVAL '70 days', NOW());
