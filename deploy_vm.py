import paramiko
import time
import os

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    sftp = ssh.open_sftp()
    
    print("Fixing go.mod files...")
    for service in ['echallan-services', 'echallan-calculator']:
        remote_go_mod = f'/home/percy/CDPI/{service}/go.mod'
        
        # Read remote go.mod
        with sftp.file(remote_go_mod, 'r') as f:
            content = f.read().decode('utf-8')
        
        # Fix go version string
        content = content.replace('go 1.25.0', 'go 1.23.0')
        content = content.replace('go 1.25', 'go 1.23.0')
        
        # Write back
        with sftp.file(remote_go_mod, 'w') as f:
            f.write(content)
            
    print("Creating cmd/echallan-calculator directory...")
    try:
        sftp.mkdir('/home/percy/CDPI/echallan-calculator/cmd')
    except IOError:
        pass
    try:
        sftp.mkdir('/home/percy/CDPI/echallan-calculator/cmd/echallan-calculator')
    except IOError:
        pass
        
    print("Uploading main.go...")
    # Upload local main.go to remote
    local_main_go = r'd:\Downloads\cdpi2\echallan-calculator\cmd\echallan-calculator\main.go'
    sftp.put(local_main_go, '/home/percy/CDPI/echallan-calculator/cmd/echallan-calculator/main.go')
    
    # Also sync repository.go deletion and challan_validator.go update!
    print("Syncing challan_validator.go...")
    local_val = r'd:\Downloads\cdpi2\echallan-services\internal\validator\challan_validator.go'
    sftp.put(local_val, '/home/percy/CDPI/echallan-services/internal/validator/challan_validator.go')
    
    print("Deleting duplicate repository.go...")
    try:
        sftp.remove('/home/percy/CDPI/echallan-calculator/internal/repository/repository.go')
    except IOError:
        pass
        
    sftp.close()
    
    print("Restarting services...")
    def run_cmd(cmd):
        print(f"--> {cmd}")
        ssh.exec_command(cmd)

    run_cmd('killall main || true')
    run_cmd('cd /home/percy/CDPI/echallan-services && nohup go run cmd/echallan-services/main.go > service.log 2>&1 &')
    run_cmd('cd /home/percy/CDPI/echallan-calculator && nohup go run cmd/echallan-calculator/main.go > calc.log 2>&1 &')
    
    print("Waiting 10s...")
    time.sleep(10)
    
    print("Checking health...")
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8079/echallan-services/health')
    print("Services:", stdout.read().decode().strip())
    
    stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8078/echallan-calculator/health')
    print("Calculator:", stdout.read().decode().strip())
    
    ssh.close()
except Exception as e:
    print("ERROR:", e)
