from linebot.v3.messaging import (
    Configuration,
    ApiClient,
    MessagingApi,
    MessagingApiBlob,
    RichMenuRequest,
)
import os


menu = {
  "size": {
    "width": 2500,
    "height": 843
  },
  "selected": True,
  "name": "Rich Menu 1",
  "chatBarText": "RichMenu",
  "areas": [
    {
      "bounds": {
        "x": 60,
        "y": 45,
        "width": 780,
        "height": 760
      },
      "action": {
        "type": "postback",
        "label": "上傳餐點",
        "data": "action=set_description"
      }
    },
    {
      "bounds": {
        "x": 860,
        "y": 45,
        "width": 780,
        "height": 760
      },
      "action": {
        "type": "postback",
        "label": "每日報告",
        "data": "action=daily_report"
      }
    },
    {
      "bounds": {
        "x": 1660,
        "y": 45,
        "width": 780,
        "height": 760
      },
      "action": {
        "type": "postback",
        "label": "加入會員",
        "data": "action=join_us"
      }
    }
  ]
}

CHANNEL_ACCESS_TOKEN = os.environ.get("CHANNEL_ACCESS_TOKEN") # Get from environment variables
configuration = Configuration(access_token=CHANNEL_ACCESS_TOKEN)
current_path = os.path.dirname(os.path.abspath(__file__))

rich_menu_request = RichMenuRequest().from_dict(menu)

with ApiClient(configuration) as api_client:
    messaging_api = MessagingApi(api_client)
    rich_menu_id_response = messaging_api.create_rich_menu(rich_menu_request)
    rich_menu_id = rich_menu_id_response.rich_menu_id
    print(f"Rich Menu ID: {rich_menu_id}")
    messaging_api_blob = MessagingApiBlob(api_client)
    with open(f'{current_path}/pic/example.png', "rb") as image_file:
        messaging_api_blob.set_rich_menu_image(
            rich_menu_id=rich_menu_id,
            body=bytearray(image_file.read()),
            _headers={'Content-Type': 'image/png'} # Adjust content type based on image format
        )
    print("Rich menu image uploaded.")
    messaging_api.set_default_rich_menu(rich_menu_id)
    print("Rich menu set as default.")
