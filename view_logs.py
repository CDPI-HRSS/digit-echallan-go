import paramiko

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    print("--- service.log ---")
    stdin, stdout, stderr = ssh.exec_command('cat /home/percy/CDPI/echallan-services/service.log')
    print(stdout.read().decode())
    
    print("--- calc.log ---")
    stdin, stdout, stderr = ssh.exec_command('cat /home/percy/CDPI/echallan-calculator/calc.log')
    print(stdout.read().decode())
    
    ssh.close()
except Exception as e:
    print("ERROR:", e)
