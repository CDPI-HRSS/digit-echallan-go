import paramiko
import time

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    go_path = '/home/percy/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.linux-amd64/bin/go'
    
    print("Restarting services with Go 1.25.0...")
    ssh.exec_command('killall main || true')
    ssh.exec_command(f'cd /home/percy/CDPI/echallan-services && nohup {go_path} run cmd/echallan-services/main.go > service.log 2>&1 &')
    ssh.exec_command(f'cd /home/percy/CDPI/echallan-calculator && nohup {go_path} run cmd/echallan-calculator/main.go > calc.log 2>&1 &')
    
    print("Waiting 15s...")
    time.sleep(15)
    
    print("Checking health...")
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8079/echallan-services/health')
    print("Services:", stdout.read().decode().strip())
    
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8078/echallan-calculator/health')
    print("Calculator:", stdout.read().decode().strip())
    
    print("--- service.log ---")
    stdin, stdout, stderr = ssh.exec_command('tail -n 10 /home/percy/CDPI/echallan-services/service.log')
    print(stdout.read().decode().strip())

    print("--- calc.log ---")
    stdin, stdout, stderr = ssh.exec_command('tail -n 10 /home/percy/CDPI/echallan-calculator/calc.log')
    print(stdout.read().decode().strip())
    
    ssh.close()
except Exception as e:
    print("ERROR:", e)
