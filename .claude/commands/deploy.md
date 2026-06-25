SSH into the Minetopia game server, pull latest changes, restart the server, and tail the logs.

Steps:
1. Run this command and stream the output to the user:
   ```
   ssh gamehost@gamehost "cd ~/minetopia/server && git -C ~/minetopia pull origin main && docker compose down && docker compose run --rm mod-sync && docker compose up -d minecraft && docker compose logs -f --tail=50 minecraft"
   ```
2. Show the output as it comes in. If the connection fails, report the exact error.
3. Once the logs are streaming, let the user know the server is up and they can press Ctrl+C to stop tailing.
4. If the user wants to stop tailing without stopping the server, run:
   ```
   ssh gamehost@gamehost "cd ~/minetopia/server && docker compose logs --tail=20 minecraft"
   ```
   to get a final snapshot instead.
