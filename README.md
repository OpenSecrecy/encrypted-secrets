# Encrypted-Secrets Kubernetes Operator

The Encrypted-Secrets Kubernetes Operator is a tool designed to enhance the security of your Kubernetes cluster by automatically decrypting encrypted secrets and storing them in Kubernetes Secrets. This operator looks for the `EncryptedSecret` custom resource kind and performs the decryption process, ensuring that sensitive information remains protected while being accessible to your applications.


## Note: `EncryptedSecrets Operator` is currently a work in progress and is in the alpha stage. Please use it with caution in production environments.
## How it Works

1. **Define Encrypted Secrets**: Create a custom resource of kind `EncryptedSecret` that includes the encrypted secret data. This custom resource can be defined using YAML or through Kubernetes API calls.

2. **Encryption Process**: The encryptedSecret custom resource contains encrypted secret data, typically using a strong encryption algorithm like AES or a supported cloud provider. The encryption process might involve using external tools or libraries.

3. **Encrypted-Secrets Operator**: The operator continuously monitors the cluster for new encryptedSecret resources.

4. **Decryption Process**: When a new encryptedSecret resource is detected, the operator initiates the decryption process using the provided decryption key or external secrets management solution.

5. **Kubernetes Secret Creation**: Once the encryptedSecret is successfully decrypted, the operator creates a Kubernetes Secret containing the decrypted secret data.

6. **Mount Secrets in Applications**: Your applications can then access the decrypted secrets by mounting the Kubernetes Secret as environment variables or files.

## Prerequisites
- Encrypted-Secrets Kubernetes Operator deployed in the cluster

## Installation (WIP)

1. Deploy the Encrypted-Secrets Operator in your Kubernetes cluster by applying the provided YAML manifests or using a package manager like Helm.

2. Verify the operator's deployment by checking the operator pod status:

   ```shell
   kubectl get pods -n <namespace>
