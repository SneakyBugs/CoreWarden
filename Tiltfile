docker_build('coredns', 'coredns')
k8s_yaml(helm('charts/coredns', name='coredns', set=["image.repository=coredns"]))
docker_build('api', '.', dockerfile="api/Dockerfile")
k8s_yaml(helm('charts/api', name='api', set=[
  "image.repository=api",
  "config.postgres.existingSecret.name=postgres-credentials",
  "config.postgres.database=development",
  ]))
