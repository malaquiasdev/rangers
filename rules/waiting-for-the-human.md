# Waiting for the human

- A turn that ends waiting for the human (a question, a decision, a login, a delivery) does two things before it ends: it calls `PushNotification`, and it creates a reminder in the Reminders list "Tarefas Claude" named "Claude espera: " followed by the action the human has to take, with the link in the body. The push alone does not reach the human's iPhone; the reminder does.
- Create the reminder with `osascript`, targeting the Reminders app and the list "Tarefas Claude". The `Stop` hook `stop-push-check` blocks the end of the turn until both exist.
- Never ask the human to complete or delete a "Claude espera: " reminder. The `UserPromptSubmit` hook `complete-reminders` marks every open reminder of the list whose name starts with "Claude espera: " as completed when the human sends the next message.
- Do not use the prefix "Claude espera: " for anything that is not a pending question to the human: it is closed automatically.
