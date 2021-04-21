# Deploy to Nomad

This is a small tool which watches for github package release events
and then updates the desired version in Nomad for a given package.
The practical upshot is that for small projects like static sites or
daemons that don't require more complicated deployment strategies,
this can handle the rollouts automatically.

For using the GitHub backend, you must export the `GITHUB_SECRET`
environment variable which must match what you have configured in your
webhook.

