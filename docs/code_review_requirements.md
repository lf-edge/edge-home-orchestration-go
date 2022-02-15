# Code review requirements

## Contents
1. [Introduction](#1-introduction)
2. [Requirements](#2-requirements)
3. [Check list](#3-check-list)

## 1. Introduction
Code review is a process in which one or more developers are systematically tasked with reviewing the code written by another developer in order to find defects and improve it. Code review should be done by project maintainers considering code quality and safety, sharing best practices and this leads to better collaboration, creating a culture of review, building team confidence in the code.

## 2. Requirements

  * Review fewer than 400 lines of code at a time
  * Take your time. Inspection rates should under 500 LOC per hour
  * Do not review for more than 60 minutes at a time
  * Set goals and capture metrics
  * Authors should annotate source code before the review
  * Use checklists
  * Establish a process for fixing defects found
  * Foster a positive code review culture
  * Embrace the subconscious implications of peer review
  * Practice lightweight code reviews

## 3. Check list
  [ ] **Requirement mistake**: A developer might have misunderstood some requirements.  
  [ ] **Hard to detect bugs**: Bugs that are related to concurrent programming (deadlock, race-condition…) are hard to detect, hard to reproduce and hard to fix.  
  [ ] **Naming**: Code must be written in compliance with the code style and naming rules in the project.  
  [ ] **Testing**: Testing can be passed automatically to a large extend. But code coverage ratio and number of tests passed are not enough. Maintainer must review the tests to make sure that the proper conditions are asserted and that tests are not overly complex.  
  [ ] **Consistency**: Often juniors on a team reinvent the wheel because they don't know about product-wide templates and libraries. It is necessary to adhere to the heritage and best practices applied in the project.  
  [ ] **Comment**: It is easy to enforce a certain ratio of comment but this is not the point with commenting. Commenting must be written carefully in proper English to explain why the code was needed in the first place and why the code was written this way.  On the other hand, code should be readable enough to avoid having comment that explains what the code do.  
  [ ] **Documentation**: Documentation is written by human and for human. A proper human review is needed to ensure that the documentation is really informative. Proper documentation is an excellent way to reduce both support cost and increase users understanding.  
  [ ] **Vulnerabilities**: There are a wide range of potential security pitfalls that require an expertise. Regular review made by a security expert is a must.
