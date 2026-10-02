.PHONY: build run test docker-up docker-down clean archive

build:
	go build -o bin/server-inventory .

run:
	go run .

test:
	go test -v ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

clean:
	rm -rf bin/ solution.zip Отчет.md

define ARCHIVE_SCRIPT
import os, zipfile, shutil
if os.path.exists('Отчёт.md') and not os.path.exists('Отчет.md'):
    shutil.copy('Отчёт.md', 'Отчет.md')
with zipfile.ZipFile('solution.zip', 'w', compression=zipfile.ZIP_DEFLATED) as z:
    for root, dirs, files in os.walk('.'):
        dirs[:] = [d for d in dirs if not d.startswith('.git') and d != 'bin']
        for f in sorted(files):
            if f in ['solution.zip', '.DS_Store'] or f.startswith('.'):
                continue
            p = os.path.join(root, f)
            z.write(p, os.path.relpath(p, '.'))
if os.path.exists('Отчет.md'):
    os.remove('Отчет.md')
print('solution.zip успешно создан с поддержкой UTF-8!')
endef
export ARCHIVE_SCRIPT

archive: clean
	@/usr/bin/python3 -c "$$ARCHIVE_SCRIPT"
