docker_build('coredns', 'coredns')
k8s_yaml(helm('charts/coredns', name='coredns', set=["image.repository=coredns"]))
