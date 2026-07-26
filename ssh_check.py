import paramiko
import time

def run_command(ssh, cmd):
    print(f"\n--- Running: {cmd} ---")
    stdin, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode('utf-8')
    err = stderr.read().decode('utf-8')
    if out: print(out)
    if err: print("ERROR:", err)
    return out

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    # 1. Pull changes
    run_command(ssh, 'cd /home/percy/CDPI && git checkout feature/updated-go-echallan && git pull origin feature/updated-go-echallan')
    
    # 2. Stop old processes
    run_command(ssh, 'killall main || true')
    
    # 3. Start echallan-services
    run_command(ssh, 'cd /home/percy/CDPI/echallan-services && nohup go run cmd/echallan-services/main.go > service.log 2>&1 &')
    
    # 4. Start echallan-calculator
    run_command(ssh, 'cd /home/percy/CDPI/echallan-calculator && nohup go run cmd/echallan-calculator/main.go > calc.log 2>&1 &')
    
    # Wait for servers to start
    print("Waiting 10 seconds for servers to start...")
    time.sleep(10)
    
    # 5. Check health
    run_command(ssh, 'curl -s http://localhost:8079/echallan-services/health || echo "Services down"')
    run_command(ssh, 'curl -s http://localhost:8078/echallan-calculator/health || echo "Calculator down"')
    
    # 6. View logs if needed
    run_command(ssh, 'tail -n 10 /home/percy/CDPI/echallan-calculator/calc.log')
    
    ssh.close()
except Exception as e:
    print("Failed to connect or execute:", e)
