# ah-helper-bot
helps to monitor AH delivery



user -> help
user -> addme       => bot: get chatId, postcode, save to db
user -> removeme    => bot: remove chatId, postcode from db

bot -> by trigger check available deliveries. If something is available, send a message

---
ah-bot internally has a channel with events.

telegram pushes an event from user
telegram listens to new message event

pubsub pushes an event for scan
pubsub listens to scan requests from external pubsub

ah-bot listens to events
when ah-bot gets an help,addme, removeme events -> message
when ah-bot gets an scan event -> message