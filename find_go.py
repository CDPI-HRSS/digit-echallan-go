import paramiko

try:
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect('20.244.110.170', username='percy', password='percy@123456', timeout=10)
    
    stdin, stdout, stderr = ssh.exec_command('ls -l /home/percy/go/bin/')
    print("go bin:", stdout.read().decode())
    
    stdin, stdout, stderr = ssh.exec_command('find /home/percy -name "go" -type f -executable 2>/dev/null')
    print("executable go:", stdout.read().decode())
    
    ssh.close()
except Exception as e:
    print("ERROR:", e)
