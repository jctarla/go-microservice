#!/bin/sh

echo "Starting..."
mkdir -p /root/.oci/
mkdir -p /root/.kube/

echo "Generating certificate..."
echo $ENV_PEM | base64 -d > /root/.oci/oci_api_key.pem
chmod 600 /root/.oci/oci_api_key.pem

echo "Creating OCI config..."
# Create the config
echo "[DEFAULT]" >> /root/.oci/config 
echo "user=$ENV_USER_OCID" >> /root/.oci/config 
echo "fingerprint=$ENV_FINGERPRINT" >> /root/.oci/config 
echo "tenancy=$ENV_TENANCY_OCID" >> /root/.oci/config 
echo "region=sa-saopaulo-1" >> /root/.oci/config 
echo "key_file=/root/.oci/oci_api_key.pem" >> /root/.oci/config 
chmod 600 /root/.oci/config 

echo "Creating OKE config..."
oci ce cluster create-kubeconfig --cluster-id $ENV_CLUSTER_ID
echo "Creating Starting APP..."
./go-microservice


