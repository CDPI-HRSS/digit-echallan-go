import paramiko
import time

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    sftp = ssh.open_sftp()
    for service in ['echallan-services', 'echallan-calculator']:
        remote_go_mod = f'/home/percy/CDPI/{service}/go.mod'
        with sftp.file(remote_go_mod, 'r') as f:
            content = f.read().decode('utf-8')
        
        content = content.replace('go 1.23.0', 'go 1.23')
        
        with sftp.file(remote_go_mod, 'w') as f:
            f.write(content)
            
    sftp.close()
    
    ssh.exec_command('killall main || true')
    ssh.exec_command('cd /home/percy/CDPI/echallan-services && nohup go run cmd/echallan-services/main.go > service.log 2>&1 &')
    ssh.exec_command('cd /home/percy/CDPI/echallan-calculator && nohup go run cmd/echallan-calculator/main.go > calc.log 2>&1 &')
    
    time.sleep(10)
    
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8079/echallan-services/health')
    print("Services:", stdout.read().decode().strip())
    
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8078/echallan-calculator/health')
    print("Calculator:", stdout.read().decode().strip())
    
    ssh.close()
except Exception as e:
    print("ERROR:", e)
