![Sample Image](https://raw.githubusercontent.com/sythe7448/EveStagingSystemRangeChecker/master/images/sample.png)

## What is it?
Eve Sonar is a tool that allows you to build a list of staging systems and compare if another system is in range. You can either do this manually or login into to Eve ESI to have it track your characters location automatically.

## Features
- Range Checking for each jump range.
- Saving staging system data to be reused each time the app is opened.
- System auto complete for inputting systems.
- For security reasons it will never store any ESI information after you close the app.
- Open source.

## Usage
Once you have the app open. You will want to make a list of staging systems in the large text field using `systemName:owner or note` and each entry/system on a new line. The system name will be validated based on eve database the owner or note can be anything you want. After you have your list of staging systems, either login to auto track or manually input systems to check the ranges.

## Contribution
Contributions are welcome! If you'd like to contribute to the project, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes.
4. Commit your changes and push them to your fork.
5. Open a pull request to the main repository.

## Bug Reporting
If you run into a bug please report it  through the project's [GitHub Issues interface](https://github.com/sythe7448/Eve-Sonar/issues). Thank you!

## To Do
- [ ] Make GitHub Actions to auto build releases.
- [ ] Improve the UI with colors and a better AutoComplete system.
- [ ] Make tests to confirm everything works without having to manually test.
- [ ] Add warning popup for the errors.
- [ ] Get the url used to validate off of local host.
- [ ] Figure out what license to use.