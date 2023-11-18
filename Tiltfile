docker_build('coredns', '.')
k8s_yaml(helm('charts/coredns', name='coredns', set=["image.repository=coredns"]))
