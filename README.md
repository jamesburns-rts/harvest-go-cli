# harvest-go-cli
CLI for everything Harvest

# Getting Started
1. Download the binary from the releases or build it yourself.
2. Go to https://id.getharvest.com/oauth2/access_tokens/new and create an access token. 
3. Copy down the `<accessToken>` and the `<accountId>`.
4. Using the binary, run:
```sh
./harvest set --harvest-access-token <accessToken>
./harvest set --harvest-account-id <accountId>
```
