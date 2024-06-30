load('ext://helm_resource', 'helm_resource', 'helm_repo')

docker_build('coredns', 'coredns')
k8s_yaml(helm('charts/coredns', name='coredns', set=["image.repository=coredns"]))

docker_build('api', '.', dockerfile="api/Dockerfile")
k8s_yaml(helm('charts/api', name='api', set=[
  "image.repository=api",
  "config.postgres.existingSecret.name=postgres-credentials",
  "config.postgres.database=development",
]))

docker_build('externaldns-webhook', '.', dockerfile="external-dns/Dockerfile")
k8s_yaml("env/manifests/webhook.yaml")
helm_repo("external-dns-repo", "https://kubernetes-sigs.github.io/external-dns/")
helm_resource(
  "external-dns",
	"external-dns-repo/external-dns",
	namespace="external-dns",
	flags=["--values=env/manifests/externaldns-values.yaml"],
	image_deps=["externaldns-webhook"],
	image_keys=[
    ("provider.webhook.image.repository", "provider.webhook.image.tag"),
	],
	resource_deps=["external-dns-repo"],
)
