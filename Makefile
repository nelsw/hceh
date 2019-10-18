# https://docs.aws.amazon.com/cli/latest/reference/lambda/index.html

.PHONY: bld pkg run it create update

#FN=hcEmailHandler
FN=$(shell yq r template.yml Resources. -j | jq 'keys[0]' -r)
ROLE=arn:aws:iam::270722761968:role/service-role/test-role
SGID=sg-0b5ae3a5e20b842bf
SNID1=subnet-49510d3f
SNID2=subnet-54056e0c
NOM=$(shell yq r template.yml Resources. -j | jq 'keys[0]' -r)

it: update

bld:
	GOOS=linux GOARCH=amd64 go build -o handler

pkg: bld
	zip handler.zip handler

run: bld
	sam local invoke -e testdata/email-confirmation.json ${FN} | jq '. | {StatusCode: .statusCode, Headers: .headers, Body: .body|fromjson}';\
#	sam local invoke -e testdata/password-reset.json ${FN} | jq '. | {StatusCode: .statusCode, Headers: .headers, Body: .body|fromjson}'

EVJ=$(shell yq r template.yml Resources.*.Properties.Environment.Variables -j | jq '.[0]' -r)
update: pkg
	aws lambda update-function-code --function-name ${FN} --zip-file fileb://./handler.zip;\
	aws lambda update-function-configuration --function-name ${FN} --environment '{ "Variables": ${EVJ} }';\

create: bld
	aws lambda create-function \
	--function-name ${FN} \
	--role ${ROLE} \
	--environment '{ "Variables": ${EVJ} }' \
	--vpc-config '{ "SubnetIds": ["${SNID1}","${SNID2}"], "SecurityGroupIds": ["${SGID}"] }'\
	--zip-file fileb://./handler.zip \
	--handler handler \
	--runtime go1.x \
	--memory-size 512 \
	--timeout 30;
