# Upload Menu

## Environment Variables

```bash
touch .env
```

In `.env`, add the following variables:

```env
CHANNEL_ACCESS_TOKEN="YOUR_CHANNEL_ACCESS_TOKEN"
CHANNEL_SECRET="YOUR_CHANNEL_SECRET"
```

## Set Menu Picture

- Put the image into `./deployment/richmenu/pic/` and name it `example.png`.

## Run Upload Menu

```bash
# Upload menu
make create-menu
```
