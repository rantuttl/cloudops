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
        stage('Deploy') {
            steps {
                sh 'make deploy'
            }
        }
    }
}
