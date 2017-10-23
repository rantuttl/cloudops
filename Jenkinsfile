pipeline {
    agent any

    stages {
	stage('Unit Test') {
	    steps {
                sh 'make test'
            }
	}
        stage('Build') {
            steps {
                sh 'make'
            }
        }
        stage('Container Image') {
            steps {
                sh 'make containers'
            }
        }
        stage('Push Image') {
            image = docker.image('cloudops-api:latest')
            docker.withRegistry('https://registry.hub.docker.com', '28252db1-5f52-4fb4-8776-041a14f362de') {
                image.push()
            }
        }
    }
}
