minikube-up:
    minikube start
    minikube addons enable ingress
    minikube addons enable registryi

minikube-reset:
    minikube delete --all
