require 'json'

package = JSON.parse(File.read(File.join(__dir__, '..', 'package.json')))

Pod::Spec.new do |s|
  s.name           = 'ExpoLighterModule'
  s.version        = package['version']
  s.summary        = package['description']
  s.description    = package['description']
  s.license        = package['license']
  s.author         = package['author']
  s.homepage       = package['homepage']
  s.platform       = :ios, '13.0'
  s.swift_version  = '5.4'
  s.source         = { git: 'https://github.com/elliottech/lighter-go' }
  s.static_framework = true

  s.dependency 'ExpoModulesCore'

  # Pod source code
  s.source_files = "**/*.{h,m,swift}"
  
  # Include the Lighter xcframework
  s.vendored_frameworks = "Lighter.xcframework"
end

