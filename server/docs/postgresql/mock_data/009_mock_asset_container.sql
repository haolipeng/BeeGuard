-- =====================================================
-- 模拟数据: asset_container (容器资产表)
-- 数据量: 60条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 容器运行时:
--   docker     - 独立 Docker 主机 (Harbor/Jenkins/GitLab/监控/ELK)
--   containerd - EKS 节点 (K8s 系统 Pod + 应用 Pod)
-- =====================================================

INSERT INTO asset_container (agent_id, host_name, host_ip, container_id, name, state, image_id, image_name, runtime, pid, create_time, created_at, updated_at) VALUES

-- ==========================================
-- aws-harbor-01 容器 (Docker, Harbor 组件)
-- host: agent-030-m7n8o9p0 / 10.0.5.12
-- ==========================================
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'a1b2c3d4e5f60011a1b2c3d4e5f60011a1b2c3d4e5f60011a1b2c3d4e5f60011', 'harbor-core', 'running', 'sha256:4a1f2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a', 'goharbor/harbor-core:v2.10.0', 'docker', '12001', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'b2c3d4e5f6a10022b2c3d4e5f6a10022b2c3d4e5f6a10022b2c3d4e5f6a10022', 'harbor-portal', 'running', 'sha256:5b2e3c4d5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b', 'goharbor/harbor-portal:v2.10.0', 'docker', '12002', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'c3d4e5f6a1b20033c3d4e5f6a1b20033c3d4e5f6a1b20033c3d4e5f6a1b20033', 'harbor-db', 'running', 'sha256:6c3f4d5e6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c', 'goharbor/harbor-db:v2.10.0', 'docker', '12003', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'd4e5f6a1b2c30044d4e5f6a1b2c30044d4e5f6a1b2c30044d4e5f6a1b2c30044', 'harbor-redis', 'running', 'sha256:7d4a5e6f7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d', 'goharbor/redis-photon:v2.10.0', 'docker', '12004', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'e5f6a1b2c3d40055e5f6a1b2c3d40055e5f6a1b2c3d40055e5f6a1b2c3d40055', 'harbor-registry', 'running', 'sha256:8e5b6f7a8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e', 'goharbor/registry-photon:v2.10.0', 'docker', '12005', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'f6a1b2c3d4e50066f6a1b2c3d4e50066f6a1b2c3d4e50066f6a1b2c3d4e50066', 'harbor-jobservice', 'running', 'sha256:9f6c7a8b9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f', 'goharbor/harbor-jobservice:v2.10.0', 'docker', '12006', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'a7b2c3d4e5f60077a7b2c3d4e5f60077a7b2c3d4e5f60077a7b2c3d4e5f60077', 'nginx', 'running', 'sha256:a07d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c', 'goharbor/nginx-photon:v2.10.0', 'docker', '12007', '2024-06-15T08:30:00Z', NOW() - INTERVAL '88 days', NOW()),

-- ==========================================
-- aws-jenkins-01 容器 (Docker, 构建代理)
-- host: agent-028-e9f0g1h2 / 10.0.5.10
-- ==========================================
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '1a2b3c4d5e6f0088112b3c4d5e6f0088112b3c4d5e6f0088112b3c4d5e6f0088', 'jenkins-agent-1', 'running', 'sha256:b18e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d', 'jenkins/inbound-agent:3206.vb_15dcf73f6a_9-3', 'docker', '23001', '2024-08-10T10:00:00Z', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '2b3c4d5e6f1a0099223c4d5e6f1a0099223c4d5e6f1a0099223c4d5e6f1a0099', 'jenkins-agent-2', 'running', 'sha256:c29f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e', 'jenkins/inbound-agent:3206.vb_15dcf73f6a_9-3', 'docker', '23002', '2024-08-10T10:05:00Z', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '3c4d5e6f1a2b00aa334d5e6f1a2b00aa334d5e6f1a2b00aa334d5e6f1a2b00aa', 'jenkins-agent-3', 'running', 'sha256:d3a01b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f', 'jenkins/inbound-agent:3206.vb_15dcf73f6a_9-3', 'docker', '23003', '2024-08-10T10:10:00Z', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '4d5e6f1a2b3c00bb445e6f1a2b3c00bb445e6f1a2b3c00bb445e6f1a2b3c00bb', 'sonarqube', 'running', 'sha256:e4b12c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a', 'sonarqube:10.4-community', 'docker', '23004', '2024-08-05T14:00:00Z', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '5e6f1a2b3c4d00cc556f1a2b3c4d00cc556f1a2b3c4d00cc556f1a2b3c4d00cc', 'sonar-db', 'running', 'sha256:f5c23d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b', 'postgres:16-alpine', 'docker', '23005', '2024-08-05T14:00:00Z', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', '6f1a2b3c4d5e00dd661a2b3c4d5e00dd661a2b3c4d5e00dd661a2b3c4d5e00dd', 'build-temp-maven', 'exited', 'sha256:a6d34e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c', 'maven:3.9-eclipse-temurin-21', 'docker', '0', '2024-12-28T15:30:00Z', NOW() - INTERVAL '2 days', NOW()),

-- ==========================================
-- aws-eks-master-01 容器 (containerd, K8s 控制面)
-- host: agent-023-k9l0m1n2 / 10.0.4.10
-- ==========================================
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'aa1122334455667788990011aabbccddeeff00112233445566778899aabb0011', 'etcd', 'running', 'sha256:1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b', 'registry.k8s.io/etcd:3.5.12-0', 'containerd', '78001', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'bb2233445566778899001122bbccddeeff00112233445566778899aabbcc0022', 'kube-apiserver', 'running', 'sha256:2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c', 'registry.k8s.io/kube-apiserver:v1.29.2', 'containerd', '78002', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'cc3344556677889900112233ccddeeff00112233445566778899aabbccdd0033', 'kube-controller-manager', 'running', 'sha256:3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d', 'registry.k8s.io/kube-controller-manager:v1.29.2', 'containerd', '78003', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'dd4455667788990011223344ddeeff00112233445566778899aabbccddeeff0044', 'kube-scheduler', 'running', 'sha256:4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e', 'registry.k8s.io/kube-scheduler:v1.29.2', 'containerd', '78004', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'ee5566778899001122334455eeff00112233445566778899aabbccddeeff110055', 'aws-vpc-cni', 'running', 'sha256:5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f', '602401143452.dkr.ecr.ap-southeast-1.amazonaws.com/amazon-k8s-cni:v1.16.0', 'containerd', '78005', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'ff6677889900112233445566ff00112233445566778899aabbccddeeff220066', 'coredns', 'running', 'sha256:6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a', 'registry.k8s.io/coredns/coredns:v1.11.1', 'containerd', '78006', '2024-09-01T09:00:00Z', NOW() - INTERVAL '60 days', NOW()),

-- ==========================================
-- aws-eks-node-01 容器 (containerd, K8s 工作节点)
-- host: agent-024-o3p4q5r6 / 10.0.4.11
-- ==========================================
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '1100aabbccddeeff2233445566778899001122334455aabbccddeeff00112201', 'kube-proxy', 'running', 'sha256:7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b', 'registry.k8s.io/kube-proxy:v1.29.2', 'containerd', '34001', '2024-09-01T09:00:00Z', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '2211bbccddeeff00334455667788990011223344556677aabbccddeeff00112202', 'aws-vpc-cni', 'running', 'sha256:8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c', '602401143452.dkr.ecr.ap-southeast-1.amazonaws.com/amazon-k8s-cni:v1.16.0', 'containerd', '34002', '2024-09-01T09:00:00Z', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '3322ccddeeff0011445566778899001122334455667788aabbccddeeff00112203', 'nginx-ingress-controller', 'running', 'sha256:9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d', 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'containerd', '34003', '2024-09-01T09:00:00Z', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '4433ddeeff001122556677889900112233445566778899aabbccddeeff00112204', 'app-frontend-7f8d9e0a1b', 'running', 'sha256:0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e', 'company/frontend:v3.2.0', 'containerd', '34004', '2024-10-10T11:30:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '5544eeff00112233667788990011223344556677889900aabbccddeeff00112205', 'app-backend-2c3d4e5f6g', 'running', 'sha256:1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f', 'company/backend:v4.1.0', 'containerd', '34005', '2024-10-10T11:30:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '6655ff0011223344778899001122334455667788990011aabbccddeeff00112206', 'prometheus-node-exporter', 'running', 'sha256:2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a', 'prom/node-exporter:v1.7.0', 'containerd', '34006', '2024-09-01T09:00:00Z', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', '7766001122334455889900112233445566778899001122aabbccddeeff00112207', 'job-backup-2024-12-28', 'exited', 'sha256:3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b', 'company/backup-job:v2.0.0', 'containerd', '0', '2024-12-28T02:00:00Z', NOW() - INTERVAL '2 days', NOW()),

-- ==========================================
-- aws-eks-node-02 容器 (containerd, K8s 工作节点)
-- host: agent-025-s7t8u9v0 / 10.0.4.12
-- ==========================================
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'aa11002233445566778899aabbccddeeff00112233445566778899001122330011', 'kube-proxy', 'running', 'sha256:7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b', 'registry.k8s.io/kube-proxy:v1.29.2', 'containerd', '45001', '2024-09-01T09:00:00Z', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'bb22113344556677889900aabbccddeeff00112233445566778899001122330022', 'aws-vpc-cni', 'running', 'sha256:8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c', '602401143452.dkr.ecr.ap-southeast-1.amazonaws.com/amazon-k8s-cni:v1.16.0', 'containerd', '45002', '2024-09-01T09:00:00Z', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'cc33224455667788990011aabbccddeeff00112233445566778899001122330033', 'app-frontend-8g9h0i1j2k', 'running', 'sha256:0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e', 'company/frontend:v3.2.0', 'containerd', '45003', '2024-10-10T11:30:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'dd44335566778899001122aabbccddeeff00112233445566778899001122330044', 'app-backend-3d4e5f6g7h', 'running', 'sha256:1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f', 'company/backend:v4.1.0', 'containerd', '45004', '2024-10-10T11:30:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'ee55446677889900112233aabbccddeeff00112233445566778899001122330055', 'redis-cache-0', 'running', 'sha256:4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b', 'redis:7.2-alpine', 'containerd', '45005', '2024-09-05T15:00:00Z', NOW() - INTERVAL '55 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'ff66557788990011223344aabbccddeeff00112233445566778899001122330066', 'mysql-primary-0', 'running', 'sha256:5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c', 'mysql:8.0.36', 'containerd', '45006', '2024-09-03T08:00:00Z', NOW() - INTERVAL '57 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'aa77668899001122334455aabbccddeeff00112233445566778899001122330077', 'cronjob-cleanup-abc123', 'exited', 'sha256:6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d', 'company/cleanup:v1.1.0', 'containerd', '0', '2024-12-29T03:00:00Z', NOW() - INTERVAL '1 day', NOW()),

-- ==========================================
-- aws-eks-node-03 容器 (containerd, K8s 工作节点)
-- host: agent-026-w1x2y3z4 / 10.0.4.13
-- ==========================================
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', '11aa22bb33cc44dd55ee66ff778899001122334455667788990011aabbccdd0011', 'kube-proxy', 'running', 'sha256:7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b', 'registry.k8s.io/kube-proxy:v1.29.2', 'containerd', '56001', '2024-09-01T09:00:00Z', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', '22bb33cc44dd55ee66ff77889900112233445566778899001122aabbccddeeff22', 'aws-vpc-cni', 'running', 'sha256:8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c', '602401143452.dkr.ecr.ap-southeast-1.amazonaws.com/amazon-k8s-cni:v1.16.0', 'containerd', '56002', '2024-09-01T09:00:00Z', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', '33cc44dd55ee66ff778899001122334455667788990011223344aabbccddeeff33', 'app-worker-1a2b3c4d5e', 'running', 'sha256:7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d', 'company/worker:v2.3.0', 'containerd', '56003', '2024-10-10T11:30:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', '44dd55ee66ff7788990011223344556677889900112233445566aabbccddeeff44', 'app-worker-4e5f6g7h8i', 'running', 'sha256:8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e', 'company/worker:v2.3.0', 'containerd', '56004', '2024-10-10T11:35:00Z', NOW() - INTERVAL '50 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', '55ee66ff778899001122334455667788990011223344556677aabbccddeeff0055', 'app-scheduler-9j0k1l2m3n', 'running', 'sha256:9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f', 'company/scheduler:v1.5.0', 'containerd', '56005', '2024-10-12T08:00:00Z', NOW() - INTERVAL '48 days', NOW()),

-- ==========================================
-- aws-eks-node-04 容器 (containerd, K8s 工作节点 - 离线)
-- host: agent-027-a5b6c7d8 / 10.0.4.14
-- ==========================================
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', '66ff778899001122334455667788990011223344556677889900aabbccddeeff66', 'kube-proxy', 'exited', 'sha256:7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b', 'registry.k8s.io/kube-proxy:v1.29.2', 'containerd', '0', '2024-09-01T09:00:00Z', NOW() - INTERVAL '52 days', NOW() - INTERVAL '1 day'),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', '7788990011223344556677889900112233445566778899001122aabbccddeeff77', 'aws-vpc-cni', 'exited', 'sha256:8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c', '602401143452.dkr.ecr.ap-southeast-1.amazonaws.com/amazon-k8s-cni:v1.16.0', 'containerd', '0', '2024-09-01T09:00:00Z', NOW() - INTERVAL '52 days', NOW() - INTERVAL '1 day'),
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', '8899001122334455667788990011223344556677889900112233aabbccddeeff88', 'app-backend-5h6i7j8k9l', 'exited', 'sha256:1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f', 'company/backend:v4.1.0', 'containerd', '0', '2024-10-10T11:30:00Z', NOW() - INTERVAL '52 days', NOW() - INTERVAL '1 day'),

-- ==========================================
-- aws-gitlab-01 容器 (Docker, GitLab Runner)
-- host: agent-029-i3j4k5l6 / 10.0.5.11
-- ==========================================
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'aabb001122334455667788990011223344556677889900112233445566778800aa', 'gitlab-runner-1', 'running', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2', 'gitlab/gitlab-runner:v16.9.1', 'docker', '89001', '2024-07-20T10:00:00Z', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'bbcc112233445566778899001122334455667788990011223344556677889900bb', 'gitlab-runner-2', 'running', 'sha256:b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3', 'gitlab/gitlab-runner:v16.9.1', 'docker', '89002', '2024-07-20T10:00:00Z', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'ccdd223344556677889900112233445566778899001122334455667788990011cc', 'gitlab-runner-3', 'running', 'sha256:c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4', 'gitlab/gitlab-runner:v16.9.1', 'docker', '89003', '2024-07-20T10:05:00Z', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'ddee334455667788990011223344556677889900112233445566778899001122dd', 'dind-service', 'running', 'sha256:d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5', 'docker:24.0-dind', 'docker', '89004', '2024-07-20T10:00:00Z', NOW() - INTERVAL '95 days', NOW()),

-- ==========================================
-- aws-prometheus-01 容器 (Docker, 监控)
-- host: agent-033-y9z0a1b2 / 10.0.6.10
-- ==========================================
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'eeff445566778899001122334455667788990011223344556677889900112233ee', 'prometheus', 'running', 'sha256:e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6', 'prom/prometheus:v2.50.0', 'docker', '94001', '2024-05-01T09:00:00Z', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'ff00556677889900112233445566778899001122334455667788990011223344ff', 'alertmanager', 'running', 'sha256:f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7', 'prom/alertmanager:v0.27.0', 'docker', '94002', '2024-05-01T09:00:00Z', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'aa11667788990011223344556677889900112233445566778899001122334455aa', 'blackbox-exporter', 'running', 'sha256:a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8', 'prom/blackbox-exporter:v0.25.0', 'docker', '94003', '2024-05-01T09:00:00Z', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'bb22778899001122334455667788990011223344556677889900112233445566bb', 'pushgateway', 'running', 'sha256:b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9', 'prom/pushgateway:v1.7.0', 'docker', '94004', '2024-05-01T09:00:00Z', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'cc33889900112233445566778899001122334455667788990011223344aabb55cc', 'thanos-sidecar', 'running', 'sha256:e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8', 'quay.io/thanos/thanos:v0.34.1', 'docker', '94005', '2024-05-01T09:00:00Z', NOW() - INTERVAL '110 days', NOW()),

-- ==========================================
-- aws-grafana-01 容器 (Docker, 可视化)
-- host: agent-034-c3d4e5f6 / 10.0.6.11
-- ==========================================
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'cc33889900112233445566778899001122334455667788990011223344556677cc', 'grafana', 'running', 'sha256:c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0', 'grafana/grafana:10.3.1', 'docker', '95001', '2024-05-15T10:00:00Z', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'dd44990011223344556677889900112233445566778899001122334455667788dd', 'grafana-renderer', 'running', 'sha256:d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1', 'grafana/grafana-image-renderer:3.9.0', 'docker', '95002', '2024-05-15T10:00:00Z', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'ee55001122334455667788990011223344556677889900112233445566778899ee', 'loki', 'running', 'sha256:e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2', 'grafana/loki:2.9.4', 'docker', '95003', '2024-05-15T10:00:00Z', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'ff66112233445566778899001122334455667788990011223344556677aabb66ff', 'tempo', 'running', 'sha256:f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3', 'grafana/tempo:2.3.1', 'docker', '95004', '2024-05-15T10:00:00Z', NOW() - INTERVAL '105 days', NOW()),

-- ==========================================
-- aws-elk-01 容器 (Docker, ELK Stack)
-- host: agent-035-g7h8i9j0 / 10.0.6.12
-- ==========================================
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'ff66112233445566778899001122334455667788990011223344556677889900ff', 'elasticsearch', 'running', 'sha256:f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3', 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 'docker', '96001', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'aa77223344556677889900112233445566778899001122334455667788990011aa', 'logstash', 'running', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4', 'docker.elastic.co/logstash/logstash:8.12.2', 'docker', '96002', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'bb88334455667788990011223344556677889900112233445566778899001122bb', 'kibana', 'running', 'sha256:b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5', 'docker.elastic.co/kibana/kibana:8.12.2', 'docker', '96003', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'cc99445566778899001122334455667788990011223344556677889900112233cc', 'filebeat', 'running', 'sha256:c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6', 'docker.elastic.co/beats/filebeat:8.12.2', 'docker', '96004', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'ddaa556677889900112233445566778899001122334455667788990011223344dd', 'apm-server', 'running', 'sha256:d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7', 'docker.elastic.co/apm/apm-server:8.12.2', 'docker', '96005', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'eebb667788990011223344556677889900112233445566778899001122334455ee', 'elastalert', 'running', 'sha256:e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8', 'jertel/elastalert2:2.17.0', 'docker', '96006', '2024-06-01T08:00:00Z', NOW() - INTERVAL '70 days', NOW());
