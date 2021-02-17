## Contributing

If you want to contribute to a project and make it better, your help is very
welcome. Contributing is also a great way to learn more about social coding on
Github, new technologies and and their ecosystems and how to make constructive,
helpful bug reports, feature requests and the noblest of all contributions:
a good, clean pull request.

You can use templates to create a description of your
[**pull request**](PULL_REQUEST_TEMPLATE.md) or [**issue**](ISSUE_TEMPLATE.md),
the design of the template will greatly simplify Edge Orchestration team work on
injecting your code. But **this is not mandatory**. We will always welcome any
help.

### How to make a clean pull request

#### 1. [Fork](http://help.github.com/fork-a-repo/) the Edge Orchestration repository on github and clone your fork to your development environment
```sh
$ git clone https://github.com/YOUR-GITHUB-USERNAME/edge-home-orchestration-go.git
```
If you have trouble setting up GIT with GitHub in Linux, or are getting errors like "Permission Denied (publickey)", then you must [setup your GIT installation to work with GitHub](http://help.github.com/linux-set-up-git/)

#### 2. Add the main Edge Orchestration repository as an additional git remote called "upstream"
Change to the directory where you cloned Edge Orchestration, normally, "Edge Orchestration". Then enter the following command:
```sh
$ git remote add upstream https://github.com/lf-edge/edge-home-orchestration-go
```

#### 3. Make sure there is an issue created for the thing you are working on.

All new features and bug fixes should have an associated issue to provide a single point of reference for discussion and documentation. Take a few minutes to look through the existing issue list for one that matches the contribution you intend to make. If you find one already on the issue list, then please leave a comment on that issue indicating you intend to work on that item. If you do not find an existing issue matching what you intend to work on, please open a new issue for your item. This will allow the team to review your suggestion, and provide appropriate feedback along the way.

> For small changes or documentation issues, you don't need to create an issue, a pull request is enough in this case.

#### 4. Fetch the latest code from the main Edge Orchestration branch
```sh
$ git fetch upstream
```
You should start at this point for every new contribution to make sure you are working on the latest code.

#### 5. Create a new branch for your feature based on the current Edge Orchestration master branch

> That's very important since you will not be able to submit more than one pull request from your account if you'll use master.

Each separate bug fix or change should go in its own branch. Branch names should be descriptive and start with the number of the issue that your code relates to. If you aren't fixing any particular issue, just skip number. For example:
```sh
$ git checkout upstream/<NAMED_RELEASE>
$ git checkout -b 999-name-of-your-branch-goes-here
```
Above, <NAMED_RELEASE> can be 'Alpha', 'Baobab', 'Coconut', etc. - see list of releases.

#### 6. Do your magic, write your code
Make sure it works and your contribution corresponds to [testing policy](../docs/testing_policy.md) :)

#### 7. Update the ReleaseNotes
Edit the ReleaseNotes file to include your change, you should insert this at the top of the file under the "Work in progress" heading, the line in the change log should look like one of the following:
```sh
Bug #999: a description of the bug fix (Your Name)
Enh #999: a description of the enhancement (Your Name)
```
`#999` is the issue number that the `Bug` or `Enh` is referring to.  
The changelog should be grouped by type (`Bug`,`Enh`) and ordered by issue number.

For very small fixes, e.g. typos and documentation changes, there is no need to update the ReleaseNotes.

#### 8. Commit your changes

add the files/changes you want to commit to the staging area with
```sh
$ git add path/to/my/file.go
```

Commit your changes with a descriptive commit message. Make sure to mention the ticket number with #XXX so github will automatically link your commit with the ticket:
```sh
$ git commit -m "A brief description of this change which fixes #42 goes here" --signoff
```

#### 9. Pull the latest Edge Orchestration code from upstream into your branch
```sh
$ git pull upstream <NAMED_RELEASE>
```
This ensures you have the latest code in your branch before you open your pull request. If there are any merge conflicts, you should fix them now and commit the changes again. This ensures that it's easy for the Edge Orchestration team to merge your changes with one click.

#### 10. Having resolved any conflicts, push your code to github
```sh
$ git push -u origin 999-name-of-your-branch-goes-here
```

The `-u` parameter ensures that your branch will now automatically push and pull from the github branch. That means if you type `git push` the next time it will know where to push to.

#### 11. Open a pull request against upstream.
Go to your repository on github and click "Pull Request", choose your branch on the right and enter some more details in the comment box. To link the pull request to the issue put anywhere in the pull comment `#999` where 999 is the issue number. <br>
_Please check out if your PR passes through an automatic build verification test provided by our CI system ([#138](https://github.com/lf-edge/edge-home-orchestration-go/pull/138))._

> Note that each pull-request should fix a single change.

#### 12. Someone will review your code
Someone will review your code, and you might be asked to make some changes, if so go to step #6 (you don't need to open another pull request if your current one is still open). If your code is accepted it will be merged into the main branch and become part of the next Edge Orchestration release. If not, don't be disheartened, different people need different features and Edge Orchestration can't be everything to everyone, your code will still be available on github as a reference for people who need it.

#### 13. Cleaning it up

After your code was either accepted or declined you can delete branches you've worked with from your local repository and `origin`.
```sh
$ git checkout <NAMED_RELEASE>
$ git branch -D 999-name-of-your-branch-goes-here
$ git push origin --delete 999-name-of-your-branch-goes-here
```
