require 'open-uri'
require 'roda'
require 'sequel'


class App < Roda
  route do |r|
    r.root do
      'Hello, world!'
    end
    r.get 'external' do
      open('http://example.com').read
    end
    r.get 'external_https' do
      open('https://example.com').read
    end
    r.get 'external_no_proxy' do
      open('http://google.com').read
    end
    r.get 'host' do
      TCPSocket.new('host.cfdev.sh', ENV['HOST_SERVER_PORT']).gets
    end
    r.get 'mysql' do
      db = Sequel.connect(ENV['DATABASE_URL'])
      "Versions: #{db.fetch('SHOW VARIABLES LIKE "%version%"').all}\n"
    end
  end
end

run App.freeze.app
