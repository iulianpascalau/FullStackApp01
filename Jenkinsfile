pipeline {
    agent any

    environment {
        PROD_VM_USER = 'ubuntu'
        PROD_VM_HOST = credentials('app-prod-vm-ip')
        PROJECT_PATH = '/home/ubuntu/app'
    }

    stages {
        stage('Deploy') {
            steps {
                sshagent(['prod-vm-ssh-key']) {
                     sh "ssh -o StrictHostKeyChecking=no ${PROD_VM_USER}@${PROD_VM_HOST} 'cd ${PROJECT_PATH} && ./scripts/deploy.sh main'"
                }
            }
        }
    }
}
