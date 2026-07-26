import paramiko
import time

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    sftp = ssh.open_sftp()
    print("Uploading updated main.go...")
    local_main_go = r'd:\Downloads\cdpi2\echallan-calculator\cmd\echallan-calculator\main.go'
    sftp.put(local_main_go, '/home/percy/CDPI/echallan-calculator/cmd/echallan-calculator/main.go')
    sftp.close()
    
    print("Restarting services...")
    ssh.exec_command('killall main || true')
    ssh.exec_command('cd /home/percy/CDPI/echallan-services && nohup go run cmd/echallan-services/main.go > service.log 2>&1 &')
    ssh.exec_command('cd /home/percy/CDPI/echallan-calculator && nohup go run cmd/echallan-calculator/main.go > calc.log 2>&1 &')
    
    print("Waiting 10s...")
    time.sleep(10)
    
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
