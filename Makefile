# https://docs.aws.amazon.com/cli/latest/reference/lambda/index.html
# https://stedolan.github.io/jq/
# https://mikefarah.github.io/yq/

.PHONY: bld pkg run it create update

FN=
ROLE=
SGID=
SNID1=
SNID2=
EVJ={ "": "" }

MJ=jq '. | {StatusCode: .statusCode, Headers: .headers, Body: .body|fromjson}'

it: create

bld:
	GOOS=linux GOARCH=amd64 go build -o handler

pkg: bld
	zip handler.zip handler

run: bld
	sam local invoke -e testdata/email-confirmation.json ${FN} | ${MJ};\
	sam local invoke -e testdata/password-reset.json ${FN} | ${MJ}

update: pkg
	aws lambda update-function-code --function-name ${FN} --zip-file fileb://./handler.zip;\
	aws lambda update-function-configuration --function-name ${FN} ;\

create: pkg
	aws lambda create-function \
	--function-name ${FN} \
	--role ${ROLE} \
	--environment '{ "Variables": ${EVJ} }' \
	--vpc-config '{ "SubnetIds": ["${SNID1}","${SNID2}"], "SecurityGroupIds": ["${SGID}"] }' \
	--zip-file fileb://./handler.zip \
	--handler handler \
	--runtime go1.x \
	--memory-size 512 \
	--timeout 30;
