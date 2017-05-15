/*
* This file send push notifications to an apple device that has the findyourhome app
* Follow the terminal arguments correctly, otherwise the push notification will not be sent correctly.
*/

var apn = require('apn');

/* Terminal Arguments:
* keyPath
* keyId
* teamId
* deviceToken
* listingLink
* area
*/
var args = process.argv.slice(2)
keyId = args[0],
teamId = args[1],
deviceToken = args[2],
listingLink = args[3],
area = args[4]

// Set up apn with the APNs Auth Key
var apnProvider = new apn.Provider({
     token: {
        key: 'apns.p8', // Path to the key p8 file
        keyId: keyId, // The Key ID of the p8 file (available at https://developer.apple.com/account/ios/certificate/key)
        teamId: teamId, // The Team ID of your Apple Developer Account (available at https://developer.apple.com/account/#/membership/)
    },
    production: false // Set to true if sending a notification to a production iOS app
});

// Prepare a new notification

var notification = new apn.Notification();

// Specify your iOS app's Bundle ID (accessible within the project editor)
notification.topic = 'com.worldyn.findyourhome';

// Set expiration to 1 hour from now (in case device is offline)
notification.expiry = Math.floor(Date.now() / 1000) + 3600;

// Set app badge indicator
notification.badge = 3;

// Play ping.aiff sound when the notification is received
notification.sound = 'ping.aiff';

// Display the following message (the actual notification text, supports emoji)
notification.alert = 'New Listing in ' + area;

// Send any extra payload data with the notification which will be accessible to your app in didReceiveRemoteNotification
notification.payload = {id: 123, area: ''};

// Actually send the notification
apnProvider.send(notification, deviceToken).then(function(result) {
    // Check the result for any failed devices
    console.log(result);
    process.exit()
});
