-- =====================================================
-- 模拟数据: vuln_info (漏洞信息表)
-- 适用环境: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- 数据量: 40条
-- 说明: 真实CVE漏洞信息，覆盖critical/high/medium/low四个等级
--       主机漏洞与容器漏洞共用此表
-- =====================================================

INSERT INTO vuln_info (id, cve_id, vuln_name, severity, cvss_score, description, fix_suggestion, reference_urls, created_at, updated_at) VALUES
-- Critical (10条)
(1,  'CVE-2021-44228', 'Apache Log4j2 远程代码执行漏洞 (Log4Shell)', 'critical', 10.0,
 'Apache Log4j2 2.0-beta9至2.14.1版本中存在JNDI注入漏洞，攻击者可通过构造恶意日志消息触发远程代码执行。',
 '升级Log4j2至2.17.1或更高版本；临时缓解：设置log4j2.formatMsgNoLookups=true',
 'https://nvd.nist.gov/vuln/detail/CVE-2021-44228',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '2 days'),

(2,  'CVE-2024-6387', 'OpenSSH regreSSHion 远程代码执行漏洞', 'critical', 8.1,
 'OpenSSH 8.5p1至9.7p1版本中sshd存在信号处理竞态条件漏洞，未经身份验证的攻击者可利用此漏洞在Linux/glibc系统上以root权限执行任意代码。',
 '升级OpenSSH至9.8p1或更高版本；临时缓解：在sshd_config中设置LoginGraceTime 0（注意可能导致DoS风险）',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-6387',
 NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 day'),

(3,  'CVE-2024-3094', 'XZ Utils 后门漏洞', 'critical', 10.0,
 'XZ Utils 5.6.0和5.6.1版本中被植入恶意后门代码，通过修改liblzma库影响OpenSSH sshd进程，可导致未经身份验证的远程代码执行。',
 '立即降级XZ Utils至5.4.x或更低安全版本；检查系统是否已被入侵',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-3094',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '3 days'),

(4,  'CVE-2021-3156', 'Sudo Buffer Overflow 提权漏洞 (Baron Samedit)', 'critical', 7.8,
 'Sudo 1.8.2至1.8.31p2以及1.9.0至1.9.5p1版本存在堆缓冲区溢出漏洞，本地用户可利用此漏洞在不需要密码的情况下获取root权限。',
 '升级Sudo至1.9.5p2或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2021-3156',
 NOW() - INTERVAL '120 days', NOW() - INTERVAL '10 days'),

(5,  'CVE-2023-44487', 'HTTP/2 Rapid Reset 拒绝服务漏洞', 'critical', 7.5,
 'HTTP/2协议实现中存在拒绝服务漏洞，攻击者可通过快速发送和取消大量HTTP/2请求流来耗尽服务器资源，已被大规模利用于DDoS攻击。',
 '升级Web服务器（nginx/Apache/Envoy等）至修复版本；配置HTTP/2最大并发流限制',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-44487',
 NOW() - INTERVAL '80 days', NOW() - INTERVAL '5 days'),

(6,  'CVE-2024-21626', 'runc 容器逃逸漏洞 (Leaky Vessels)', 'critical', 8.6,
 'runc 1.1.11及之前版本中存在文件描述符泄露漏洞，攻击者可利用WORKDIR指令通过/proc/self/fd/引用宿主机文件系统，实现容器逃逸。',
 '升级runc至1.1.12或更高版本；升级Docker至24.0.9或25.0.2',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-21626',
 NOW() - INTERVAL '50 days', NOW() - INTERVAL '2 days'),

(7,  'CVE-2023-38545', 'curl SOCKS5 堆缓冲区溢出漏洞', 'critical', 9.8,
 'curl 7.69.0至8.3.0版本中SOCKS5代理握手过程存在堆缓冲区溢出漏洞，攻击者可通过恶意SOCKS5代理服务器触发远程代码执行。',
 '升级curl至8.4.0或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-38545',
 NOW() - INTERVAL '100 days', NOW() - INTERVAL '8 days'),

(8,  'CVE-2023-4911', 'glibc ld.so 本地提权漏洞 (Looney Tunables)', 'critical', 7.8,
 'GNU C Library (glibc) 2.34至2.38版本中动态加载器ld.so在处理GLIBC_TUNABLES环境变量时存在缓冲区溢出漏洞，本地攻击者可利用SUID程序获取root权限。',
 '升级glibc至修复版本；临时缓解：使用SystemTap脚本阻止设置GLIBC_TUNABLES',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-4911',
 NOW() - INTERVAL '95 days', NOW() - INTERVAL '6 days'),

(9,  'CVE-2024-1086', 'Linux内核netfilter nf_tables本地提权漏洞', 'critical', 7.8,
 'Linux内核5.14至6.6版本中netfilter nf_tables子系统存在use-after-free漏洞，本地攻击者可利用nft_verdict_init()函数中的整数下溢实现本地提权。',
 '升级Linux内核至修复版本（6.1.76+/6.6.15+/6.7.3+）',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-1086',
 NOW() - INTERVAL '55 days', NOW() - INTERVAL '4 days'),

(10, 'CVE-2023-32233', 'Linux内核nf_tables UAF本地提权漏洞', 'critical', 7.8,
 'Linux内核6.3.1之前版本中Netfilter nf_tables在处理匿名集合时存在use-after-free漏洞，本地用户可利用此漏洞进行权限提升。',
 '升级Linux内核至6.3.2或更高版本；临时缓解：限制非特权用户创建用户命名空间',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-32233',
 NOW() - INTERVAL '110 days', NOW() - INTERVAL '12 days'),

-- High (12条)
(11, 'CVE-2023-5528', 'Kubernetes Windows节点权限提升漏洞', 'high', 7.2,
 'Kubernetes 1.28.3及之前版本中存在权限提升漏洞，可通过创建使用本地卷的Pod在Windows节点上以SYSTEM权限执行命令。',
 '升级Kubernetes至1.28.4或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-5528',
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '7 days'),

(12, 'CVE-2023-46604', 'Apache ActiveMQ 远程代码执行漏洞', 'high', 7.5,
 'Apache ActiveMQ 5.18.3之前版本中OpenWire协议序列化类类型过滤器存在缺陷，远程攻击者可利用此漏洞在代理端执行任意Shell命令。',
 '升级ActiveMQ至5.18.3或5.17.6',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-46604',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '9 days'),

(13, 'CVE-2023-3446', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3,
 'OpenSSL 1.0.2至3.1.1版本中DH_check()函数在检查过大的DH参数时耗时过长，攻击者可利用此漏洞造成拒绝服务。',
 '升级OpenSSL至3.1.2/3.0.10/1.1.1v或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-3446',
 NOW() - INTERVAL '105 days', NOW() - INTERVAL '15 days'),

(14, 'CVE-2023-5363', 'OpenSSL密钥和IV长度处理漏洞', 'high', 7.5,
 'OpenSSL 3.0.0至3.1.3版本中EVP_EncryptInit_ex2/EVP_DecryptInit_ex2/EVP_CipherInit_ex2在设置过长密钥或IV时处理不当，可能导致截断或溢出。',
 '升级OpenSSL至3.1.4/3.0.12或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-5363',
 NOW() - INTERVAL '88 days', NOW() - INTERVAL '11 days'),

(15, 'CVE-2023-39325', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5,
 'Go 1.21.3之前版本的net/http和golang.org/x/net/http2中对HTTP/2请求处理存在拒绝服务漏洞，与CVE-2023-44487相关。',
 '升级Go至1.21.3/1.20.10或更高版本；升级golang.org/x/net至0.17.0',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-39325',
 NOW() - INTERVAL '82 days', NOW() - INTERVAL '5 days'),

(16, 'CVE-2023-44270', 'PostCSS 换行符解析漏洞', 'high', 5.3,
 'PostCSS 8.4.31之前版本在解析外部CSS中的换行符时存在缺陷，攻击者可利用此漏洞绕过Linter插件实现CSS注入。',
 '升级PostCSS至8.4.31或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-44270',
 NOW() - INTERVAL '75 days', NOW() - INTERVAL '8 days'),

(17, 'CVE-2023-45853', 'zlib MiniZip 整数溢出漏洞', 'high', 9.8,
 'zlib 1.3之前版本的MiniZip中zipOpenNewFileInZip4_64函数存在整数溢出漏洞，可能导致堆缓冲区溢出。',
 '升级zlib至1.3或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-45853',
 NOW() - INTERVAL '92 days', NOW() - INTERVAL '7 days'),

(18, 'CVE-2023-2650', 'OpenSSL ASN.1对象标识符处理DoS', 'high', 6.5,
 'OpenSSL 1.0.2至3.1.0版本在处理某些ASN.1对象标识符时存在性能问题，可被利用进行拒绝服务攻击。',
 '升级OpenSSL至3.1.1/3.0.9/1.1.1u或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-2650',
 NOW() - INTERVAL '115 days', NOW() - INTERVAL '14 days'),

(19, 'CVE-2024-0567', 'GnuTLS 证书链验证漏洞', 'high', 7.5,
 'GnuTLS 3.8.3之前版本在验证证书链中的Cockpit证书时存在漏洞，攻击者可利用构造的证书触发拒绝服务。',
 '升级GnuTLS至3.8.3或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-0567',
 NOW() - INTERVAL '48 days', NOW() - INTERVAL '3 days'),

(20, 'CVE-2023-6246', 'glibc __fortify_fail 本地提权漏洞', 'high', 7.8,
 'GNU C Library 2.37及之前版本的syslog函数中存在堆缓冲区溢出漏洞，本地攻击者可通过构造恶意输入进行权限提升。',
 '升级glibc至2.39或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-6246',
 NOW() - INTERVAL '52 days', NOW() - INTERVAL '4 days'),

(21, 'CVE-2023-48795', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9,
 'OpenSSH 9.6之前版本中SSH传输协议存在前缀截断攻击漏洞(Terrapin Attack)，攻击者可在MITM场景下降低连接安全性。',
 '升级OpenSSH至9.6或更高版本；禁用受影响的加密算法chacha20-poly1305和CBC with ETM MAC',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-48795',
 NOW() - INTERVAL '65 days', NOW() - INTERVAL '2 days'),

(22, 'CVE-2023-6129', 'OpenSSL POLY1305 MAC 计算漏洞', 'high', 6.5,
 'OpenSSL 3.0.0至3.1.4版本中POLY1305 MAC在PowerPC CPU上的向量寄存器处理存在缺陷，可能导致应用程序崩溃或产生不正确的MAC结果。',
 '升级OpenSSL至3.1.5/3.0.13或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-6129',
 NOW() - INTERVAL '58 days', NOW() - INTERVAL '5 days'),

-- Medium (10条)
(23, 'CVE-2023-5678', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3,
 'OpenSSL 1.0.2至3.1.4版本中DH密钥生成在使用过大的Q参数值时耗时过长，可被利用进行拒绝服务。',
 '升级OpenSSL至3.1.5/3.0.13/1.1.1x或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-5678',
 NOW() - INTERVAL '72 days', NOW() - INTERVAL '6 days'),

(24, 'CVE-2023-4806', 'glibc getaddrinfo() UAF漏洞', 'medium', 5.9,
 'GNU C Library 2.36及之前版本的getaddrinfo()函数在处理/etc/nsswitch.conf中的SUCCESS=continue时存在use-after-free漏洞。',
 '升级glibc至修复版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-4806',
 NOW() - INTERVAL '98 days', NOW() - INTERVAL '13 days'),

(25, 'CVE-2023-39326', 'Go net/http 请求体读取漏洞', 'medium', 5.3,
 'Go 1.21.5之前版本的net/http中HTTP请求体读取存在漏洞，恶意HTTP客户端可读取超出声明的Content-Length数据。',
 '升级Go至1.21.5/1.20.12或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-39326',
 NOW() - INTERVAL '68 days', NOW() - INTERVAL '4 days'),

(26, 'CVE-2023-45287', 'Go crypto/tls RSA密钥交换时序泄露漏洞', 'medium', 5.3,
 'Go 1.20之前版本的crypto/tls中RSA密钥交换存在可观察的时序差异，理论上攻击者可利用精确时序侧信道恢复明文。',
 '升级Go至1.20或更高版本；禁用RSA密钥交换',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-45287',
 NOW() - INTERVAL '78 days', NOW() - INTERVAL '9 days'),

(27, 'CVE-2023-4527', 'glibc getaddrinfo() 栈缓冲区溢出', 'medium', 6.5,
 'GNU C Library 2.36及之前版本的getaddrinfo()函数在处理overlong的/etc/resolv.conf中的no-aaaa选项时存在栈缓冲区读取溢出。',
 '升级glibc至修复版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-4527',
 NOW() - INTERVAL '100 days', NOW() - INTERVAL '16 days'),

(28, 'CVE-2023-52425', 'libexpat XML解析DoS漏洞', 'medium', 5.5,
 'libexpat 2.6.0之前版本在解析大型XML文档中大量嵌套实体引用时存在性能问题，可导致CPU资源耗尽。',
 '升级libexpat至2.6.0或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-52425',
 NOW() - INTERVAL '42 days', NOW() - INTERVAL '2 days'),

(29, 'CVE-2024-0727', 'OpenSSL PKCS12解析空指针解引用', 'medium', 5.5,
 'OpenSSL在解析格式错误的PKCS12文件时可能产生空指针解引用，导致应用程序崩溃。',
 '升级OpenSSL至3.2.1/3.1.5/3.0.13/1.1.1x或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2024-0727',
 NOW() - INTERVAL '46 days', NOW() - INTERVAL '3 days'),

(30, 'CVE-2023-5981', 'GnuTLS RSA-PSK时序侧信道漏洞', 'medium', 5.9,
 'GnuTLS 3.8.2之前版本在RSA-PSK密钥交换中存在时序侧信道漏洞，攻击者可在MITM场景下恢复明文。',
 '升级GnuTLS至3.8.2或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-5981',
 NOW() - INTERVAL '88 days', NOW() - INTERVAL '10 days'),

(31, 'CVE-2023-39318', 'Go html/template 脚本注入漏洞', 'medium', 6.1,
 'Go 1.21.1之前版本的html/template包在处理HTML类属性中的JavaScript操作时存在不当处理，可能导致跨站脚本攻击。',
 '升级Go至1.21.1/1.20.8或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-39318',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '11 days'),

(32, 'CVE-2023-29406', 'Go net/http Host头注入漏洞', 'medium', 6.5,
 'Go 1.20.6之前版本的net/http中HTTP客户端在发送Host头时未充分验证，可能导致头注入攻击。',
 '升级Go至1.20.6/1.19.11或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-29406',
 NOW() - INTERVAL '108 days', NOW() - INTERVAL '18 days'),

-- Low (8条)
(33, 'CVE-2023-5156', 'glibc getaddrinfo() 内存泄漏', 'low', 3.7,
 'GNU C Library 2.34至2.38版本中getaddrinfo()函数在特定条件下存在内存泄漏，长期运行的程序可能因此耗尽内存。',
 '升级glibc至修复版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-5156',
 NOW() - INTERVAL '95 days', NOW() - INTERVAL '14 days'),

(34, 'CVE-2023-6237', 'OpenSSL RSA解密性能漏洞', 'low', 3.7,
 'OpenSSL 3.0.0至3.1.4版本在使用过大的RSA密钥进行解密时可能耗时过长，但由于利用条件苛刻实际影响有限。',
 '升级OpenSSL至3.1.5/3.0.13或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-6237',
 NOW() - INTERVAL '56 days', NOW() - INTERVAL '5 days'),

(35, 'CVE-2023-4016', 'procps-ng ps命令栈缓冲区溢出', 'low', 3.3,
 'procps-ng 4.0.3之前版本的ps命令在处理精心构造的用户名时存在栈缓冲区溢出，但实际利用难度较高。',
 '升级procps-ng至4.0.4或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-4016',
 NOW() - INTERVAL '110 days', NOW() - INTERVAL '20 days'),

(36, 'CVE-2023-50387', 'DNSSEC KeyTrap 拒绝服务漏洞', 'low', 3.1,
 'DNSSEC验证解析器在处理包含大量DNSKEY记录的响应时存在CPU资源消耗问题，攻击者可通过恶意DNS区域触发拒绝服务。',
 '升级DNS解析器软件（BIND/Unbound等）至修复版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-50387',
 NOW() - INTERVAL '40 days', NOW() - INTERVAL '2 days'),

(37, 'CVE-2023-45803', 'Python urllib3 请求体泄露漏洞', 'low', 4.2,
 'urllib3 2.0.7之前版本在HTTP 303重定向后未正确移除请求体，可能导致敏感数据泄露至非预期的URL。',
 '升级urllib3至2.0.7/1.26.18或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-45803',
 NOW() - INTERVAL '76 days', NOW() - INTERVAL '8 days'),

(38, 'CVE-2023-44271', 'Pillow 图像解析DoS漏洞', 'low', 3.3,
 'Pillow 10.0.1之前版本在解析特定图像文件时内存使用不当，攻击者可通过恶意图像导致内存耗尽。',
 '升级Pillow至10.0.1或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-44271',
 NOW() - INTERVAL '82 days', NOW() - INTERVAL '12 days'),

(39, 'CVE-2023-2975', 'OpenSSL AES-SIV空关联数据处理漏洞', 'low', 3.7,
 'OpenSSL 3.0.0至3.1.1版本中AES-SIV加密在处理空关联数据时存在逻辑错误，导致认证失效但实际影响有限。',
 '升级OpenSSL至3.1.2/3.0.10或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-2975',
 NOW() - INTERVAL '112 days', NOW() - INTERVAL '22 days'),

(40, 'CVE-2023-4641', 'shadow-utils useradd密码信息泄露', 'low', 3.3,
 'shadow-utils 4.14.0之前版本的useradd命令在特定条件下可能将密码信息写入/etc/shadow时权限设置不当。',
 '升级shadow-utils至4.14.0或更高版本',
 'https://nvd.nist.gov/vuln/detail/CVE-2023-4641',
 NOW() - INTERVAL '105 days', NOW() - INTERVAL '17 days');

-- 重置序列
SELECT setval('vuln_info_id_seq', 40);
