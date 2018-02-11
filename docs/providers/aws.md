## Amazon Web Services

zbuild can be configured to use AWS. Internally, it uses S3 and DynamoDB to maintain metadata and artifact information.

### AWS Prerequisites

A little bit of AWS setup needs to be done. At a minimum, you will need a role that can create buckets, create Dynamo DB tables, and read/write from these buckets and tables. A slightly better approach is to have three roles.

#### 0. Named Profiles

See the [named profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) section of the AWS documentation for information on how to setup named profiles locally.

#### 1. Setup Role

The role you need for setting up zbuild is a role that can create an S3 bucket and Dynamo DB tables. This role is only required during initial setup.

#### 2. Developer Role

The developer role only needs to permissions to read from the S3 bucket and from the Dynamo tables.

#### 3. Publisher Role

The publisher role needs read and write access to the S3 bucket and the Dynamo DB tables. Depending on the size of your team and your desired layout, it may make sense to merge this with the developer role. If you use a build server, you'll want to make this role available to the build server so it can publish artifacts.

## Initialize the workspace

The `init-workspace` command will ask you for a some configuration parameters. These parameters are used to initialize a local workspace, but you will have an option to create the AWS resources at the end of the setup. The configuration parameters are:

* **Source set name**: this is the name of the source set for the workspace.
* **Backend type**: you must select **AWS** here.
* **S3 Bucket**: The name of the S3 bucket to store artifacts in. This are shared across all source sets and will be created if it doesn't exist.
* **Artifact Table Name**: Used to store artifact metadata in Dynamo DB (default: zbuild-artifact-metadata)
* **Source Set Table Name**: Used to store metadata about source sets (default: zbuild-source-set-metadata)
* **Region**: The region for Dynamo DB
* **Profile**: The name of the credentials profile

At the end you will be prompted to create the resources.

Now you're all set! Your team mates will need to go through this prompt themselves, using the same AWS resource names you used. They won't need to create the resources, however.
