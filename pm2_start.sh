#/bin/bash
pm2 delete MCTAS || echo ""
pm2 start ./bin/linux/amd64/MastersClosetServer --name MCTAS -- config_local.json
pm2 save
pm2 log MCTAS