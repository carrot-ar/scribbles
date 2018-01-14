echo "==1== building image"
docker build --no-cache -t gcr.io/senior-project-proving-grounds/scribbles .
echo "==2== pushing image to gcr"
gcloud docker -- push gcr.io/senior-project-proving-grounds/scribbles
