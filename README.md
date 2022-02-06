# k8s-controller-demo

Learning to create a custom kubernetes controller following https://book.kubebuilder.io/introduction.html.

But this controller wraps papermill(jupyter notebook execution) around the built in CronJob to execute notebooks on schedule and reducing all the CronJob jargon with the built in API.


Sample custom resource definition
```yaml
apiVersion: papermill.papermill.dev/v1
kind: ScheduledNotebook
metadata:
  name: schedulednotebook-sample
spec:
  name: andersontesting
  dockerImage: test-app
  schedule: "*/1 * * * *"
  inputNotebook: "local:///home/app/input.ipynb"
  outputNotebook: "local:///home/app/out.ipynb"
  parameters:
    name: Anderson
```
