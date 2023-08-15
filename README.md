# simple-csi-driver
Initialy I just want to add nfs support, from csi suggests, nfs is not required to add a controllerpublishvolume, but I will add it in case it is needed.


csi-sanity --ginkgo.v --csi.testvolumeparameters="${ROOT_DIR}/test/sanity/sanity-params.yaml" --csi.endpoint="unix://${ROOT_DIR}/csi.sock"