# Converts XML documents exported from WordPress to individual Markdown files
# for use with scriptogr.am.
#
# The XML document should be placed in the same directory as this script and
# be given the name "source.xml".  Output files will be created in a folder
# called "md".  Each file will be named after its post date.
#
# Titles, tags, post dates and content is preserved.  Pages are not exported.
# Unpublished pages are exported and "published" states are not preserved.
# Posts that contain HTML are automatically converted to Markdown format.  It
# is possible that HTML-formatted posts contain formatting errors; their publish
# dates are printed by the script so that the user can manually fix any issues
# before re-running the script.
#
# Usage: python3 wptosgrm.py

import xml.etree.ElementTree as etree
import cgi
import shutil
import os
import html2text
import sys

class Comment:
	def __init__(self):
		self.date = ""
		self.author = ""
		self.authorEmail = ""
		self.text = ""
		self.filename = ""

class Post:
	def __init__(self):
		self.title = ""
		self.name = ""
		self.date = ""
		self.link = ""
		self.text = ""
		self.filename = ""
		self.author = ""
		self.id = ""
		self.tags = []
		self.comments = []

def load(file):
	xml = etree.parse(file)
	root = xml.getroot()
	return root.find("channel").findall("item")

outputPath = "./md"
inputPath = "./source.xml"

if os.path.exists(outputPath):
	shutil.rmtree(outputPath)

os.makedirs(outputPath)

items = load(inputPath)

for item in items:

	if item.find("{http://wordpress.org/export/1.2/}post_type").text != "post":
		continue

	if item.find("{http://wordpress.org/export/1.2/}status").text != "publish":
		continue

	post = Post()

	post.id = item.find("{http://wordpress.org/export/1.2/}post_id").text
	post.name = item.find("{http://wordpress.org/export/1.2/}post_name").text
	post.title = item.find("title").text
	post.link = item.find("link").text
	post.date = item.find("{http://wordpress.org/export/1.2/}post_date_gmt").text
	post.text = item.find("{http://purl.org/rss/1.0/modules/content/}encoded").text
	post.filename = post.date.replace(' ', '_').replace(':', '-')

	# Only run the post through the HTML to Markdown parser if it contains HTML
	# tags.
	if ">" in post.text:
		try:
			post.text = html2text.html2text(post.text)
		except:

			# Sometimes we get malformed HTML (or just tags that the parser
			# can't handle).  We just print the name of the file so that it
			# can be fixed manually.
			print(post.filename)

	for category in item.findall("category"):
		if category.attrib["domain"] != "post_tag":
			continue

		post.tags.append(category.text)


	for xmlComment in item.findall("{http://wordpress.org/export/1.2/}comment"):
		comment = Comment()

		comment.author = xmlComment.find("{http://wordpress.org/export/1.2/}comment_author").text
		comment.authorEmail = xmlComment.find("{http://wordpress.org/export/1.2/}comment_author_email").text
		comment.date = xmlComment.find("{http://wordpress.org/export/1.2/}comment_date_gmt").text
		comment.text = xmlComment.find("{http://wordpress.org/export/1.2/}comment_content").text
		comment.filename = comment.date.replace(' ', '_').replace(':', '-')

		post.comments.append(comment)

	if os.path.exists(os.path.join(outputPath, post.name + ".md")):
		sys.exit("Duplicate post name")

	output = open(os.path.join(outputPath, post.name + ".md"), 'wt')
	output.write("Title: " + post.title + "\n")
	output.write("Date: " + post.date + "\n")
	output.write("Tags: " + ", ".join(post.tags))
	output.write("\n\n")
	output.write(post.text)
	output.close()

	if len(post.comments) > 0:

		postPath = os.path.join(outputPath, post.name)

		if os.path.exists(postPath):
			shutil.rmtree(postPath)

		os.makedirs(postPath)

		commentPath = os.path.join(postPath, "comments")

		if os.path.exists(commentPath):
			shutil.rmtree(commentPath)

		os.makedirs(commentPath)

		for comment in post.comments:

			if os.path.exists(os.path.join(commentPath, comment.filename + ".md")):
				sys.exit("Duplicate comment name")

			output = open(os.path.join(commentPath, comment.filename + ".md"), 'wt')
			output.write("Author: " + comment.author + "\n")

			if (comment.authorEmail):
				output.write("Email: " + comment.authorEmail + "\n")

			output.write("Date: " + comment.date + "\n")
			output.write("\n\n")
			output.write(comment.text)
			output.close()

