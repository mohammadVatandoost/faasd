# Bootstrap faasd on Digitalocean

1) [Sign up to DigitalOcean](https://www.digitalocean.com/?refcode=2962aa9e56a1&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=CopyPaste)
2) [Download Terraform](https://www.terraform.io)
3) Clone this gist using the URL from the address bar
4) Run `terraform init`
5) Run `terraform apply -var="do_token=$(cat $HOME/digitalocean-access-token)"`
6) View the output for the login command and gateway URL i.e.

```
gateway_url = http://178.128.39.201:8080/
login_cmd = faas-cli login -g http://178.128.39.201:8080/ -p rvIU49CEcFcHmqxj
password = rvIU49CEcFcHmqxj
```

Note that the user-data may take a couple of minutes to come up since it will be pulling in various components and preparing the machine.

A single host with 1GB of RAM will be deployed for you, to remove at a later date simply use `terraform destroy`.

If required, you can remove the VM via `terraform destroy -var="do_token=$(cat $HOME/digitalocean-access-token)"`
