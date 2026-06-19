# Upload LINE Menu

## 1. Prerequisites

- Install [Docker](https://docs.docker.com/desktop/)

## 2. Configure Environment Variables

- Create a .env file in the root directory and add the following variables in the .env file:

```env
CHANNEL_ACCESS_TOKEN="YOUR_CHANNEL_ACCESS_TOKEN"
CHANNEL_SECRET="YOUR_CHANNEL_SECRET"
```

## 3. Set Menu Picture

- Put the image into `./deployment/richmenu/pic/` and name it `example.png`.

## 4. Run Docker

- If you already have a menu picture on your LINE OA, you need to delete it first. Use the following command to delete the menu:

```bash
make delete-menu-docker
```

or you can use the following command to delete the menu without CMake:

```bash
docker build -t delete-menu-app -f ./deployment/richmenu/delete.Dockerfile .
docker run --rm --env-file .env delete-menu-app
```

- If you don't have a menu picture on your LINE OA, you can skip the above step.

- Use the following command to upload the menu:

```bash
make create-menu-docker
```

or you can use the following command to upload the menu without CMake:

```bash
docker build -t richmenu-app -f ./deployment/richmenu/create.Dockerfile .
docker run --rm --env-file .env richmenu-app
```
