key_label=$1
key_string=$2

#key_string=$(cat /dev/urandom | xxd -p | dd bs=32 count=1 2>/dev/null)
#key_string=$(cat /dev/urandom | base64 | sed 's/+//g' | sed 's/\///g' | tr -d '\n' | dd count=1 bs=32 2>/dev/null)

js_string="db.apiKeys.insert({keyString: \"${key_string}\", enabled: true, label: \"${key_label}\"});"

echo $js_string | mongo crawler

#echo $key_string
