require 'open-uri'
require 'roda'
require 'sequel'

DB = Sequel.connect(ENV['DATABASE_URL'])

class App < Roda
  route do |r|
    r.root do
      'Hello, world!'
    end
    r.get 'external' do
      open('http://example.com').read
    end
    r.get 'host' do
      TCPSocket.new('host.cfdev.sh', ENV['HOST_SERVER_PORT']).gets
    end
    r.get 'mysql' do
      "Versions: #{DB.fetch('SHOW VARIABLES LIKE "%version%"').all}\n"
    end
  end
end

run App.freeze.app
