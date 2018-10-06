# s3-flood

Creates many small files in an S3 bucket.

### Environment Variables

* `REGION` - The S3 region where the bucket is located
* `BUCKET` - The bucket where the files are to be written
* `FILE_PREFIX` - The prefix used for all files written
* `FILE_COUNT` - How many files should be written
* `CONCURRENCY` - How many simultaneous writes should be performed
