# ecs-deploy

This is an app to make it easier to update the image of already created ECS tasks.


### Installation
Download a release and put the binary on your path.

### To Use
Once the binary is on your systems path you can just call 
```
> ecs-deploy <options>
```

#### Command line Arguments
Currently supported
```
  -cluster string
        ECS Cluster (default "cluster")
  -container string
        container to update (for use with multi container tasks)
  -count int
        Number of desired -1 means ignored (default -1)
  -credFile string
        Full path to credentials file (default "C:\\Users\\chris\\.aws\\credentials")
  -family string
        Task Family - if not provided service name is used
  -image string
        Image to run (default "image")
  -noWait
        Disable waiting
  -profile string
        Credential Profile To Use (default "default")
  -region string
        AWS Region (default "us-east-1")
  -service string
        ECS Service to update (default "service")
  -taskName string
        Task ARN to update (optional)
```

### Development

#### Tests

### Docker

### TODO
- Allow mem and cpu updates
- If only one image make container not needed
- Tests

License
----

MIT
